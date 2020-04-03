package control

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxive/os/config"
	"github.com/maxive/os/pkg/log"
	"github.com/maxive/os/pkg/util"
)

func BootstrapMain() {
	log.InitLogger()

	log.Debugf("bootstrapAction")
	if err := UdevSettle(); err != nil {
		log.Errorf("Failed to run udev settle: %v", err)
	}

	log.Debugf("bootstrapAction: loadingConfig")
	cfg := config.LoadConfig()

	log.Debugf("bootstrapAction: Rngd(%v)", cfg.Maxive.State.Rngd)
	if cfg.Maxive.State.Rngd {
		if err := runRngd(); err != nil {
			log.Errorf("Failed to run rngd: %v", err)
		}
	}

	log.Debugf("bootstrapAction: MdadmScan(%v)", cfg.Maxive.State.MdadmScan)
	if cfg.Maxive.State.MdadmScan {
		if err := mdadmScan(); err != nil {
			log.Errorf("Failed to run mdadm scan: %v", err)
		}
	}

	log.Debugf("bootstrapAction: cryptsetup(%v)", cfg.Maxive.State.Cryptsetup)
	if cfg.Maxive.State.Cryptsetup {
		if err := cryptsetup(); err != nil {
			log.Errorf("Failed to run cryptsetup: %v", err)
		}
	}

	log.Debugf("bootstrapAction: LvmScan(%v)", cfg.Maxive.State.LvmScan)
	if cfg.Maxive.State.LvmScan {
		if err := vgchange(); err != nil {
			log.Errorf("Failed to run vgchange: %v", err)
		}
	}

	stateScript := cfg.Maxive.State.Script
	log.Debugf("bootstrapAction: stateScript(%v)", stateScript)
	if stateScript != "" {
		if err := runStateScript(stateScript); err != nil {
			log.Errorf("Failed to run state script: %v", err)
		}
	}

	log.Debugf("bootstrapAction: RunCommandSequence(%v)", cfg.Bootcmd)
	err := util.RunCommandSequence(cfg.Bootcmd)
	if err != nil {
		log.Error(err)
	}

	if cfg.Maxive.State.Dev != "" && cfg.Maxive.State.Wait {
		waitForRoot(cfg)
	}

	if len(cfg.Maxive.State.Autoformat) > 0 {
		log.Infof("bootstrap container: Autoformat(%v) as %s", cfg.Maxive.State.Autoformat, "ext4")
		if err := autoformat(cfg.Maxive.State.Autoformat); err != nil {
			log.Errorf("Failed to run autoformat: %v", err)
		}
	}

	log.Debugf("bootstrapAction: udev settle2")
	if err := UdevSettle(); err != nil {
		log.Errorf("Failed to run udev settle: %v", err)
	}
}

func mdadmScan() error {
	cmd := exec.Command("mdadm", "--assemble", "--scan")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func vgchange() error {
	cmd := exec.Command("vgchange", "--activate", "ay")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func cryptsetup() error {
	devices, err := util.BlkidType("crypto_LUKS")
	if err != nil {
		return err
	}

	for _, cryptdevice := range devices {
		fdRead, err := os.Open("/dev/console")
		if err != nil {
			return err
		}
		defer fdRead.Close()

		fdWrite, err := os.OpenFile("/dev/console", os.O_WRONLY|os.O_APPEND, 0)
		if err != nil {
			return err
		}
		defer fdWrite.Close()

		cmd := exec.Command("cryptsetup", "luksOpen", cryptdevice, fmt.Sprintf("luks-%s", filepath.Base(cryptdevice)))
		cmd.Stdout = fdWrite
		cmd.Stderr = fdWrite
		cmd.Stdin = fdRead

		if err := cmd.Run(); err != nil {
			log.Errorf("Failed to run cryptsetup for %s: %v", cryptdevice, err)
		}
	}

	return nil
}

func runRngd() error {
	// use /dev/urandom as random number input for rngd
	// this is a really bad idea
	// since I am simple filling the kernel entropy pool with entropy coming from the kernel itself!
	// but this does not need to consider the user's hw rngd drivers.
	cmd := exec.Command("rngd", "-r", "/dev/urandom", "-q")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runStateScript(script string) error {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	if _, err := f.WriteString(script); err != nil {
		return err
	}
	if err := f.Chmod(os.ModePerm); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return util.RunScript(f.Name())
}

func waitForRoot(cfg *config.CloudConfig) {
	var dev string
	for i := 0; i < 30; i++ {
		dev = util.ResolveDevice(cfg.Maxive.State.Dev)
		if dev != "" {
			break
		}
		time.Sleep(time.Millisecond * 1000)
	}
	if dev == "" {
		return
	}
	for i := 0; i < 30; i++ {
		if _, err := os.Stat(dev); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 1000)
	}
}

func autoformat(autoformatDevices []string) error {
	cmd := exec.Command("/usr/sbin/auto-format.sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = []string{
		"AUTOFORMAT=" + strings.Join(autoformatDevices, " "),
	}
	return cmd.Run()
}
