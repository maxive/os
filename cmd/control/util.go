package control

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/log"

	"github.com/pkg/errors"
)

func yes(question string) bool {
	fmt.Printf("%s [y/N]: ", question)
	in := bufio.NewReader(os.Stdin)
	line, err := in.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}

	return strings.ToLower(line[0:1]) == "y"
}

func formatImage(image string, cfg *config.CloudConfig) string {
	domainRegistry := cfg.Maxive.Environment["REGISTRY_DOMAIN"]
	if domainRegistry != "docker.io" && domainRegistry != "" {
		return fmt.Sprintf("%s/%s", domainRegistry, image)
	}
	return image
}

func symLinkEngineBinary() []symlink {
	baseSymlink := []symlink{
		{"/usr/share/ros/os-release", "/usr/lib/os-release"},
		{"/usr/share/ros/os-release", "/etc/os-release"},

		{"/var/lib/maxive/engine/docker", "/usr/bin/docker"},
		{"/var/lib/maxive/engine/dockerd", "/usr/bin/dockerd"},
		{"/var/lib/maxive/engine/docker-init", "/usr/bin/docker-init"},
		{"/var/lib/maxive/engine/docker-proxy", "/usr/bin/docker-proxy"},

		// >= 18.09.0
		{"/var/lib/maxive/engine/containerd", "/usr/bin/containerd"},
		{"/var/lib/maxive/engine/ctr", "/usr/bin/ctr"},
		{"/var/lib/maxive/engine/containerd-shim", "/usr/bin/containerd-shim"},
		{"/var/lib/maxive/engine/runc", "/usr/bin/runc"},

		// < 18.09.0
		{"/var/lib/maxive/engine/docker-containerd", "/usr/bin/docker-containerd"},
		{"/var/lib/maxive/engine/docker-containerd-ctr", "/usr/bin/docker-containerd-ctr"},
		{"/var/lib/maxive/engine/docker-containerd-shim", "/usr/bin/docker-containerd-shim"},
		{"/var/lib/maxive/engine/docker-runc", "/usr/bin/docker-runc"},
	}
	return baseSymlink
}

func checkZfsBackingFS(driver, dir string) error {
	if driver != "zfs" {
		return nil
	}
	for i := 0; i < 4; i++ {
		mountInfo, err := ioutil.ReadFile("/proc/self/mountinfo")
		if err != nil {
			continue
		}
		for _, mount := range strings.Split(string(mountInfo), "\n") {
			if strings.Contains(mount, dir) && strings.Contains(mount, driver) {
				return nil
			}
		}
		time.Sleep(1 * time.Second)
	}
	return errors.Errorf("BackingFS: %s not match storage-driver: %s", dir, driver)
}

func checkGlobalCfg() bool {
	_, err := os.Stat("/proc/1/root/boot/global.cfg")
	if err == nil || os.IsExist(err) {
		return true
	}
	return false
}
