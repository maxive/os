package b2d

import (
	"os"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/init/configfiles"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/util"
)

const (
	boot2DockerMagic string = "boot2docker, please format-me"
)

var (
	boot2DockerEnvironment bool
)

func B2D(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	if _, err := os.Stat("/var/lib/boot2docker"); os.IsNotExist(err) {
		err := os.Mkdir("/var/lib/boot2docker", 0755)
		if err != nil {
			log.Errorf("Failed to create boot2docker directory: %v", err)
		}
	}

	if dev := util.ResolveDevice("LABEL=B2D_STATE"); dev != "" {
		boot2DockerEnvironment = true
		cfg.Maxive.State.Dev = "LABEL=B2D_STATE"
		log.Infof("boot2DockerEnvironment %s: %s", cfg.Maxive.State.Dev, dev)
		return cfg, nil
	}

	devices := []string{"/dev/sda", "/dev/vda"}
	data := make([]byte, len(boot2DockerMagic))

	for _, device := range devices {
		f, err := os.Open(device)
		if err == nil {
			defer f.Close()

			_, err = f.Read(data)
			if err == nil && string(data) == boot2DockerMagic {
				boot2DockerEnvironment = true
				cfg.Maxive.State.Dev = "LABEL=B2D_STATE"
				cfg.Maxive.State.Autoformat = []string{device}
				log.Infof("boot2DockerEnvironment %s: Autoformat %s", cfg.Maxive.State.Dev, cfg.Maxive.State.Autoformat[0])

				break
			}
		}
	}

	// save here so the bootstrap service can see it (when booting from iso, its very early)
	if boot2DockerEnvironment {
		if err := config.Set("maxive.state.dev", cfg.Maxive.State.Dev); err != nil {
			log.Errorf("Failed to update maxive.state.dev: %v", err)
		}
		if err := config.Set("maxive.state.autoformat", cfg.Maxive.State.Autoformat); err != nil {
			log.Errorf("Failed to update maxive.state.autoformat: %v", err)
		}
	}

	return config.LoadConfig(), nil
}

func Env(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	log.Debugf("memory Resolve.conf == [%s]", configfiles.ConfigFiles["/etc/resolv.conf"])

	if boot2DockerEnvironment {
		if err := config.Set("maxive.state.dev", cfg.Maxive.State.Dev); err != nil {
			log.Errorf("Failed to update maxive.state.dev: %v", err)
		}
		if err := config.Set("maxive.state.autoformat", cfg.Maxive.State.Autoformat); err != nil {
			log.Errorf("Failed to update maxive.state.autoformat: %v", err)
		}
	}

	return config.LoadConfig(), nil
}
