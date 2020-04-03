package compose

import (
	"fmt"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/docker"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/util/network"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
)

func LoadService(p *project.Project, cfg *config.CloudConfig, useNetwork bool, service string) error {
	// First check the multi engine service file.
	// If the name has been found in multi enging service file and matches, will not execute network.LoadServiceResource
	// Otherwise will execute network.LoadServiceResource
	bytes, err := network.LoadMultiEngineResource(service)
	if err != nil || bytes == nil {
		bytes, err = network.LoadServiceResource(service, useNetwork, cfg)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	m := map[interface{}]interface{}{}
	if err = yaml.Unmarshal(bytes, &m); err != nil {
		e := fmt.Errorf("Failed to parse YAML configuration for %s: %v", service, err)
		log.Error(e)
		return e
	}

	m = adjustContainerNames(m)

	bytes, err = yaml.Marshal(m)
	if err != nil {
		e := fmt.Errorf("Failed to marshal YAML configuration for %s: %v", service, err)
		log.Error(e)
		return e
	}

	if err = p.Load(bytes); err != nil {
		e := fmt.Errorf("Failed to load %s: %v", service, err)
		log.Error(e)
		return e
	}

	return nil
}

func LoadSpecialService(p *project.Project, cfg *config.CloudConfig, serviceName, serviceValue string) error {
	// Save config in case load fails
	previousConfig, ok := p.ServiceConfigs.Get(serviceName)

	p.ServiceConfigs.Add(serviceName, &composeConfig.ServiceConfig{})

	if err := LoadService(p, cfg, true, serviceValue); err != nil {
		// Rollback to previous config
		if ok {
			p.ServiceConfigs.Add(serviceName, previousConfig)
		}
		return err
	}

	return nil
}

func loadConsoleService(cfg *config.CloudConfig, p *project.Project) error {
	if cfg.Maxive.Console == "" || cfg.Maxive.Console == "default" {
		return nil
	}
	return LoadSpecialService(p, cfg, "console", cfg.Maxive.Console)
}

func loadEngineService(cfg *config.CloudConfig, p *project.Project) error {
	if cfg.Maxive.Docker.Engine == "" || cfg.Maxive.Docker.Engine == cfg.Maxive.Defaults.Docker.Engine {
		return nil
	}
	return LoadSpecialService(p, cfg, "docker", cfg.Maxive.Docker.Engine)
}

func projectReload(p *project.Project, useNetwork *bool, loadConsole bool, environmentLookup *docker.ConfigEnvironment, authLookup *docker.ConfigAuthLookup) func() error {
	enabled := map[interface{}]interface{}{}
	return func() error {
		cfg := config.LoadConfig()

		environmentLookup.SetConfig(cfg)
		authLookup.SetConfig(cfg)

		enabled = addServices(p, enabled, cfg.Maxive.Services)

		for service, serviceEnabled := range cfg.Maxive.ServicesInclude {
			if _, ok := enabled[service]; ok || !serviceEnabled {
				continue
			}

			if err := LoadService(p, cfg, *useNetwork, service); err != nil {
				if err != network.ErrNoNetwork {
					log.Errorf("Failed to load service(%s): %v", service, err)
				}
				continue
			}

			enabled[service] = service
		}

		if !*useNetwork {
			return nil
		}

		if loadConsole {
			if err := loadConsoleService(cfg, p); err != nil {
				log.Errorf("Failed to load maxive.console=(%s): %v", cfg.Maxive.Console, err)
			}
		}

		if err := loadEngineService(cfg, p); err != nil {
			log.Errorf("Failed to load maxive.docker.engine=(%s): %v", cfg.Maxive.Docker.Engine, err)
		}

		return nil
	}
}
