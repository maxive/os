package bootstrap

import (
	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/compose"
	"github.com/maxive/os/pkg/init/docker"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/sysinit"
	"github.com/maxive/os/pkg/util"
)

func bootstrapServices(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if util.ResolveDevice(cfg.Maxive.State.Dev) != "" && len(cfg.Bootcmd) == 0 {
		log.Info("NOT Running Bootstrap")

		return cfg, nil
	}
	log.Info("Running Bootstrap")
	_, err := compose.RunServiceSet("bootstrap", cfg, cfg.Maxive.BootstrapContainers)
	return cfg, err
}

func Bootstrap(cfg *config.CloudConfig) error {
	log.Info("Launching Bootstrap Docker")

	c, err := docker.Start(cfg)
	if err != nil {
		return err
	}

	defer docker.Stop(c)

	_, err = config.ChainCfgFuncs(cfg,
		[]config.CfgFuncData{
			{"bootstrap loadImages", sysinit.LoadBootstrapImages},
			{"bootstrap Services", bootstrapServices},
		})
	return err
}
