package sharedroot

import (
	"os"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/init/fsmount"

	"github.com/docker/docker/pkg/mount"
)

func Setup(c *config.CloudConfig) (*config.CloudConfig, error) {
	if c.Maxive.NoSharedRoot {
		return c, nil
	}

	if fsmount.IsInitrd() {
		for _, i := range []string{"/mnt", "/media", "/var/lib/system-docker"} {
			if err := os.MkdirAll(i, 0755); err != nil {
				return c, err
			}
			if err := mount.Mount("tmpfs", i, "tmpfs", "rw"); err != nil {
				return c, err
			}
			if err := mount.MakeShared(i); err != nil {
				return c, err
			}
		}
		return c, nil
	}

	return c, mount.MakeShared("/")
}
