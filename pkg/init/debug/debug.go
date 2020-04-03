package debug

import (
	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/log"
)

func PrintAndLoadConfig(_ *config.CloudConfig) (*config.CloudConfig, error) {
	PrintConfig()

	cfg := config.LoadConfig()
	return cfg, nil
}

func PrintConfig() {
	cfgString, err := config.Export(false, true)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Error serializing config")
	} else {
		log.Debugf("Config: %s", cfgString)
	}
}
