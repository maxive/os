// +build linux

package init

import (
	"fmt"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/dfs"
	"github.com/maxive/os/pkg/init/b2d"
	"github.com/maxive/os/pkg/init/cloudinit"
	"github.com/maxive/os/pkg/init/configfiles"
	"github.com/maxive/os/pkg/init/debug"
	"github.com/maxive/os/pkg/init/docker"
	"github.com/maxive/os/pkg/init/env"
	"github.com/maxive/os/pkg/init/fsmount"
	"github.com/maxive/os/pkg/init/hypervisor"
	"github.com/maxive/os/pkg/init/modules"
	"github.com/maxive/os/pkg/init/one"
	"github.com/maxive/os/pkg/init/prepare"
	"github.com/maxive/os/pkg/init/recovery"
	"github.com/maxive/os/pkg/init/selinux"
	"github.com/maxive/os/pkg/init/sharedroot"
	"github.com/maxive/os/pkg/init/switchroot"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/sysinit"
)

func MainInit() {
	log.InitLogger()
	// TODO: this breaks and does nothing if the cfg is invalid (or is it due to threading?)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Starting Recovery console: %v\n", r)
			recovery.Recovery(nil)
		}
	}()

	if err := RunInit(); err != nil {
		log.Fatal(err)
	}
}

func RunInit() error {
	initFuncs := config.CfgFuncs{
		{"set env", env.Init},
		{"preparefs", prepare.FS},
		{"save init cmdline", prepare.SaveCmdline},
		{"mount OEM", fsmount.MountOem},
		{"debug save cfg", debug.PrintAndLoadConfig},
		{"load modules", modules.LoadModules},
		{"recovery console", recovery.LoadRecoveryConsole},
		{"b2d env", b2d.B2D},
		{"mount STATE and bootstrap", fsmount.MountStateAndBootstrap},
		{"cloud-init", cloudinit.CloudInit},
		{"read cfg and log files", configfiles.ReadConfigFiles},
		{"switchroot", switchroot.SwitchRoot},
		{"mount OEM2", fsmount.MountOem},
		{"mount BOOT", fsmount.MountBoot},
		{"write cfg and log files", configfiles.WriteConfigFiles},
		{"b2d Env", b2d.Env},
		{"hypervisor tools", hypervisor.Tools},
		{"preparefs2", prepare.FS},
		{"load modules2", modules.LoadModules},
		{"set proxy env", env.Proxy},
		{"init SELinux", selinux.Initialize},
		{"setupSharedRoot", sharedroot.Setup},
		{"sysinit", sysinit.RunSysInit},
	}

	cfg, err := config.ChainCfgFuncs(nil, initFuncs)
	if err != nil {
		recovery.Recovery(err)
	}

	launchConfig, args := docker.GetLaunchConfig(cfg, &cfg.Maxive.SystemDocker)
	launchConfig.Fork = !cfg.Maxive.SystemDocker.Exec
	//launchConfig.NoLog = true

	log.Info("Launching System Docker")
	_, err = dfs.LaunchDocker(launchConfig, config.SystemDockerBin, args...)
	if err != nil {
		log.Errorf("Error Launching System Docker: %s", err)
		recovery.Recovery(err)
		return err
	}
	// Code never gets here - maxive.system_docker.exec=true

	return one.PidOne()
}
