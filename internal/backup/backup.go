package backup

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"prostic/internal/config"
	"prostic/internal/util"
	"strings"
	"time"
)

var log = util.GroupLogger("BACKUP")

func RunBackup() error {
	if config.Get() == nil {
		return errors.New("no config provided")
	}
	log.Infof("Check restic config")
	if !isResticConfigCorrect() {
		log.Info(config.Get().Restic)
		return errors.New("restic config invalid. Maybe you need to init the repo. Restic init")
	}
	backupID := randomID(10)
	for _, vm := range config.Get().VMs {
		err := runVMBackup(vm, backupID)
		if err != nil {
			return fmt.Errorf("could not backup vm %d: %v", vm.ID, err)
		}
	}
	log.Infof("Backup completed successfully")
	return nil
}

func runVMBackup(vm config.VM, backupID string) error {
	log.Infof("Start backup for id %d", vm.ID)
	vmPrefix := "lxc"
	if vm.IsVM {
		vmPrefix = "vm"
	}
	today := time.Now().Format("2006-01-02")
	for _, disk := range vm.Disks {
		parts := strings.Split(disk, "/")
		lvName := parts[len(parts)-1]
		snapName := lvName + "-snap"
		snapPath := filepath.Join(filepath.Dir(disk), snapName)
		destFile := filepath.Join(fmt.Sprintf("%s-%d", vmPrefix, vm.ID), strings.Split(disk, "/")[2], lvName+".raw")
		if _, err := os.Stat(disk + "-snap"); err == nil {
			log.Infof("Snapshot %s already exists", disk+"-snap. Delete it")
			if err := removeSnapshot(snapPath); err != nil {
				return err
			}
		}
		cmd := exec.Command("/usr/sbin/lvcreate", "-s", "-n", snapName, "-L", "5G", disk)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create snapshot for %s: %v\n%s", disk, err, string(out))
		}
		log.Infof("Created snapshot for %s", disk)
		err = util.RunResticCommand(true, "backup",
			"--stdin-from-command",
			fmt.Sprintf("--stdin-filename=%s", destFile),
			"--tag", fmt.Sprintf("vm=%d", vm.ID),
			"--tag", fmt.Sprintf("destFile=%s", destFile),
			"--tag", fmt.Sprintf("srcFile=%s", disk),
			"--tag", fmt.Sprintf("id=%s", backupID),
			"--tag", fmt.Sprintf("vmtype=%s", vmPrefix),
			"--tag", fmt.Sprintf("name=%s", vm.Name),
			"--tag", "type=disk",
			"--tag", fmt.Sprintf("date=%s", today),
			"--",
			"/bin/dd", "if="+snapPath, "bs=4M")
		if err != nil {
			removeSnapshot(snapPath)
			return fmt.Errorf("restic backup failed for %s: %v", snapPath, err)
		}
		log.Infof("Disk %s", disk)
		if err := removeSnapshot(snapPath); err != nil {
			return err
		}
	}

	var srcConfig string
	if vm.IsVM {
		srcConfig = filepath.Join("/etc/pve/qemu-server", fmt.Sprintf("%d.conf", vm.ID))
	} else {
		srcConfig = filepath.Join("/etc/pve/lxc", fmt.Sprintf("%d.conf", vm.ID))
	}
	destConfig := filepath.Join(fmt.Sprintf("%s-%d", vmPrefix, vm.ID), "config")
	if _, err := os.Stat(srcConfig); err == nil {
		err = util.RunResticCommand(true, "backup",
			"--stdin-from-command",
			fmt.Sprintf("--stdin-filename=%s", destConfig),
			"--tag", fmt.Sprintf("vm=%d", vm.ID),
			"--tag", fmt.Sprintf("destFile=%s", destConfig),
			"--tag", fmt.Sprintf("srcFile=%s", srcConfig),
			"--tag", "type=config",
			"--tag", fmt.Sprintf("id=%s", backupID),
			"--tag", fmt.Sprintf("date=%s", today),
			"--tag", fmt.Sprintf("vmtype=%s", vmPrefix),
			"--tag", fmt.Sprintf("name=%s", vm.Name),
			"--",
			"/bin/dd", "if="+srcConfig, "bs=4M")
		log.Infof("Config copied to %s", destConfig)
	} else {
		log.Warnf("Config file %s not found, skipping", srcConfig)
	}
	return nil
}

func removeSnapshot(path string) error {
	cmd := exec.Command("/usr/sbin/lvremove", "-f", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove snapshot %s: %v\n%s", path, err, string(out))
	}
	log.Infof("Snapshot %s removed", path)
	return nil
}

func isResticConfigCorrect() bool {
	err := util.RunResticCommand(false, "snapshots")
	return err == nil
}

const charset = "0123456789abcdefghijklmnopqrstuvwxyz"

func randomID(n int) string {
	id := make([]byte, n)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}
