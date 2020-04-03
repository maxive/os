package recovery

import (
	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/compose"
	"github.com/maxive/os/pkg/init/docker"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/netconf"
	"github.com/maxive/os/pkg/sysinit"

	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/yaml"
)

var (
	// TODO: move this into the os-config file so it can be customised.
	recoveryDockerService = composeConfig.ServiceConfigV1{
		Image: config.OsBase,
		Command: yaml.Command{
			"ros",
			"recovery-init",
		},
		Labels: map[string]string{
			config.DetachLabel: "false",
			config.ScopeLabel:  "system",
		},
		LogDriver:  "json-file",
		Net:        "host",
		Uts:        "host",
		Pid:        "host",
		Ipc:        "host",
		Privileged: true,
		Volumes: []string{
			"/dev:/host/dev",
			"/etc/ssl/certs/ca-certificates.crt:/etc/ssl/certs/ca-certificates.crt.maxive",
			"/lib/modules:/lib/modules",
			"/lib/firmware:/lib/firmware",
			"/usr/bin/ros:/usr/bin/ros:ro",
			"/usr/bin/ros:/usr/bin/cloud-init-save",
			"/usr/bin/ros:/usr/bin/respawn:ro",
			"/usr/share/ros:/usr/share/ros:ro",
			"/var/lib/maxive:/var/lib/maxive",
			"/var/lib/maxive/conf:/var/lib/maxive/conf",
			"/var/run:/var/run",
		},
	}
)

func LoadRecoveryConsole(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if cfg.Maxive.Recovery {
		Recovery(nil)
	}
	return cfg, nil
}

func Recovery(initFailure error) {
	if initFailure != nil {
		log.Errorf("MaxiveOS has failed to boot: %v", initFailure)
	}
	log.Info("Launching recovery console")

	var recoveryConfig config.CloudConfig
	recoveryConfig.Maxive.Defaults = config.Defaults{
		Network: netconf.NetworkConfig{
			DNS: netconf.DNSConfig{
				Nameservers: []string{
					"8.8.8.8",
					"8.8.4.4",
				},
			},
		},
	}
	recoveryConfig.Maxive.BootstrapDocker = config.DockerConfig{
		EngineOpts: config.EngineOpts{
			Bridge:        "none",
			StorageDriver: "overlay2",
			Restart:       &[]bool{false}[0],
			Graph:         "/var/lib/recovery-docker",
			Group:         "root",
			Host:          []string{"unix:///var/run/system-docker.sock"},
			UserlandProxy: &[]bool{false}[0],
		},
	}

	_, err := docker.Start(&recoveryConfig)
	if err != nil {
		log.Fatal(err)
	}

	_, err = config.ChainCfgFuncs(&recoveryConfig,
		[]config.CfgFuncData{
			{"loadSystemImages", sysinit.LoadBootstrapImages},
			{"recovery console", recoveryServices},
		})
	if err != nil {
		log.Fatal(err)
	}
}

func recoveryServices(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	_, err := compose.RunServiceSet("recovery", cfg, map[string]*composeConfig.ServiceConfigV1{
		"recovery": &recoveryDockerService,
	})
	return nil, err
}
