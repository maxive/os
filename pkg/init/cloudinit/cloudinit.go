package cloudinit

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/compose"
	"github.com/maxive/os/pkg/init/docker"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/sysinit"
	"github.com/maxive/os/pkg/util"
)

func CloudInit(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	stateConfig := config.LoadConfigWithPrefix(config.StateDir)
	cfg.Maxive.CloudInit.Datasources = stateConfig.Maxive.CloudInit.Datasources

	hypervisor := util.GetHypervisor()
	if hypervisor == "" {
		log.Infof("ros init: No Detected Hypervisor")
	} else {
		log.Infof("ros init: Detected Hypervisor: %s", hypervisor)
	}
	if hypervisor == "vmware" {
		// add vmware to the end - we don't want to over-ride an choices the user has made
		cfg.Maxive.CloudInit.Datasources = append(cfg.Maxive.CloudInit.Datasources, hypervisor)
	}

	exoscale, err := onlyExoscale()
	if err != nil {
		log.Error(err)
	}
	if exoscale {
		cfg.Maxive.CloudInit.Datasources = append([]string{"exoscale"}, cfg.Maxive.CloudInit.Datasources...)
	}

	proxmox, err := onlyProxmox()
	if err != nil {
		log.Error(err)
	}
	if proxmox {
		cfg.Maxive.CloudInit.Datasources = append([]string{"proxmox"}, cfg.Maxive.CloudInit.Datasources...)
	}

	if len(cfg.Maxive.CloudInit.Datasources) == 0 {
		log.Info("No specific datasources, ignore cloudinit")
		return cfg, nil
	}
	if onlyConfigDrive(cfg.Maxive.CloudInit.Datasources) {
		configDev := util.ResolveDevice("LABEL=config-2")
		if configDev == "" {
			// Check v9fs: https://www.kernel.org/doc/Documentation/filesystems/9p.txt
			matches, _ := filepath.Glob("/sys/bus/virtio/drivers/9pnet_virtio/virtio*/mount_tag")
			if len(matches) == 0 {
				log.Info("Configdrive was enabled but has no configdrive device or filesystem, ignore cloudinit")
				return cfg, nil
			}
		}
	}

	if err := config.Set("maxive.cloud_init.datasources", cfg.Maxive.CloudInit.Datasources); err != nil {
		log.Error(err)
	}

	if stateConfig.Maxive.Network.DHCPTimeout > 0 {
		cfg.Maxive.Network.DHCPTimeout = stateConfig.Maxive.Network.DHCPTimeout
		if err := config.Set("maxive.network.dhcp_timeout", stateConfig.Maxive.Network.DHCPTimeout); err != nil {
			log.Error(err)
		}
	}

	if len(stateConfig.Maxive.Network.WifiNetworks) > 0 {
		cfg.Maxive.Network.WifiNetworks = stateConfig.Maxive.Network.WifiNetworks
		if err := config.Set("maxive.network.wifi_networks", stateConfig.Maxive.Network.WifiNetworks); err != nil {
			log.Error(err)
		}
	}

	if len(stateConfig.Maxive.Network.Interfaces) > 0 {
		// DO also uses static networking, but this IP may change if:
		// 1. not using Floating IP
		// 2. creating a droplet with a snapshot, the snapshot cached the previous IP
		if onlyDigitalOcean(cfg.Maxive.CloudInit.Datasources) {
			log.Info("Do not use the previous network settings on DigitalOcean")
		} else {
			cfg.Maxive.Network = stateConfig.Maxive.Network
			if err := config.Set("maxive.network", stateConfig.Maxive.Network); err != nil {
				log.Error(err)
			}
		}
	}

	log.Infof("init, runCloudInitServices(%v)", cfg.Maxive.CloudInit.Datasources)
	if err := runCloudInitServices(cfg); err != nil {
		log.Error(err)
	}

	// It'd be nice to push to rsyslog before this, but we don't have network
	log.AddRSyslogHook()

	return config.LoadConfig(), nil
}

func runCloudInitServices(cfg *config.CloudConfig) error {
	c, err := docker.Start(cfg)
	if err != nil {
		return err
	}

	defer docker.Stop(c)

	_, err = config.ChainCfgFuncs(cfg,
		[]config.CfgFuncData{
			{"cloudinit loadImages", sysinit.LoadBootstrapImages},
			{"cloudinit Services", runCloudInitServiceSet},
		})
	return err
}

func runCloudInitServiceSet(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	log.Info("Running cloud-init services")
	_, err := compose.RunServiceSet("cloud-init", cfg, cfg.Maxive.CloudInitServices)
	return cfg, err
}

func onlyConfigDrive(datasources []string) bool {
	if len(datasources) != 1 {
		return false
	}
	for _, ds := range datasources {
		parts := strings.SplitN(ds, ":", 2)
		if parts[0] == "configdrive" {
			return true
		}
	}
	return false
}

func onlyDigitalOcean(datasources []string) bool {
	if len(datasources) != 1 {
		return false
	}
	for _, ds := range datasources {
		parts := strings.SplitN(ds, ":", 2)
		if parts[0] == "digitalocean" {
			return true
		}
	}
	return false
}

func onlyExoscale() (bool, error) {
	f, err := ioutil.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(string(f), "Exoscale"), nil
}

func onlyProxmox() (bool, error) {
	f, err := ioutil.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		return false, err
	}

	return strings.Contains(string(f), "Proxmox"), nil
}
