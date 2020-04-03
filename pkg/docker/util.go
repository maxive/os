package docker

import (
	"github.com/maxive/os/config"

	composeConfig "github.com/docker/libcompose/config"
)

func IsSystemContainer(serviceConfig *composeConfig.ServiceConfig) bool {
	return serviceConfig.Labels[config.ScopeLabel] == config.System
}
