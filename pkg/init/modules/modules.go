package modules

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/util"
)

func LoadModules(cfg *config.CloudConfig) (*config.CloudConfig, error) {
	mounted := map[string]bool{}

	f, err := os.Open("/proc/modules")
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	reader := bufio.NewScanner(f)
	for reader.Scan() {
		mounted[strings.SplitN(reader.Text(), " ", 2)[0]] = true
	}

	if util.GetHypervisor() == "hyperv" {
		cfg.Maxive.Modules = append(cfg.Maxive.Modules, "hv_utils", "hv_storvsc", "hv_vmbus")
	}

	for _, module := range cfg.Maxive.Modules {
		if mounted[module] {
			continue
		}

		log.Debugf("Loading module %s", module)
		// split module and module parameters
		cmdParam := strings.SplitN(module, " ", -1)
		cmd := exec.Command("modprobe", cmdParam...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Errorf("Could not load module %s, err %v", module, err)
		}
	}

	return cfg, reader.Err()
}
