package hypervisor

import (
	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/util"
)

func Tools(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	enableHypervisorService(cfg, util.GetHypervisor())
	return config.LoadConfig(), nil
}

func enableHypervisorService(cfg *config.CloudConfig, hypervisorName string) {
	if hypervisorName == "" {
		return
	}

	// enable open-vm-tools and hyperv-vm-tools
	// these services(xenhvm-vm-tools, kvm-vm-tools, and bhyve-vm-tools) don't exist yet
	serviceName := ""
	switch hypervisorName {
	case "vmware":
		serviceName = "open-vm-tools"
	case "hyperv":
		serviceName = "hyperv-vm-tools"
	default:
		log.Infof("no hypervisor matched")
	}

	if serviceName != "" {
		if !cfg.Maxive.HypervisorService {
			log.Infof("Skipping %s as `maxive.hypervisor_service` is set to false", serviceName)
			return
		}

		// Check removed - there's an x509 cert failure on first boot of an installed system
		// check quickly to see if there is a yml file available
		//	if service.ValidService(serviceName, cfg) {
		log.Infof("Setting maxive.services_include. %s=true", serviceName)
		if err := config.Set("maxive.services_include."+serviceName, "true"); err != nil {
			log.Error(err)
		}
	}
}
