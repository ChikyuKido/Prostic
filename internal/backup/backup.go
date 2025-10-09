package backup

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"prostic/internal/config"
	"prostic/internal/util"
	"strconv"
	"strings"
	"syscall"
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
	log.Infof("Check backup dir")
	if _, err := os.Stat(config.Get().Backup.Dir); err != nil {
		log.Infof("Backup dir does not exist. Create it")
		err := os.MkdirAll(config.Get().Backup.Dir, 0755)
		if err != nil {
			return errors.New("backup dir could not be created")
		}
	}
	enough, err := isSpaceSufficient()
	if err != nil {
		return fmt.Errorf("could not check space sufficient: %v", err)
	}
	if !enough {
		return errors.New("backup dir does not have enough space")
	}
	for _, vm := range config.Get().VMs {
		err = runVMBackup(vm)
		if err != nil {
			return fmt.Errorf("could not backup vm %d: %v", vm.ID, err)
		}
	}
	log.Infof("Finished backuping the vm's to local storage. Will not upload to restic")
	if err := util.RunResticCommand(true, "backup", config.Get().Backup.Dir); err != nil {
		return fmt.Errorf("could not backup restic: %v", err)
	}
	log.Infof("Backup completed successfully")
	return nil
}

func runVMBackup(vm config.VM) error {
	log.Infof("Start backup for id %d", vm.ID)
	backupDir := config.Get().Backup.Dir
	vmPrefix := "lxc"
	if vm.IsVM {
		vmPrefix = "vm"
	}
	vmDir := filepath.Join(backupDir, fmt.Sprintf("%s-%d", vmPrefix, vm.ID))
	if _, err := os.Stat(vmDir); err == nil {
		log.Infof("A backup already exists at %s. Delete it", vmDir)
		err = os.RemoveAll(vmDir)
		if err != nil {
			log.Warnf("Could not remove backup directory %s. This isn't critical the files will be overwritten: %v", vmDir, err)
		}
	}
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory %s: %v", vmDir, err)
	}

	for _, disk := range vm.Disks {
		parts := strings.Split(disk, "/")
		lvName := parts[len(parts)-1]
		snapName := lvName + "-snap"

		cmd := exec.Command("lvcreate", "-s", "-n", snapName, "-L", "5G", disk)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create snapshot for %s: %v\n%s", disk, err, string(out))
		}
		log.Infof("Created snapshot for %s", disk)

		snapPath := filepath.Join("/dev", "pve", snapName)
		destFile := filepath.Join(vmDir, lvName+".raw")

		dd := exec.Command("dd", "if="+snapPath, "of="+destFile, "bs=1M", "status=progress", "conv=sparse")
		dd.Stdout = os.Stdout
		dd.Stderr = os.Stderr
		log.Infof("Starting dd copy of %s to %s", snapPath, destFile)
		if err := dd.Run(); err != nil {
			removeSnapshot(snapName)
			return fmt.Errorf("dd failed for %s: %v", snapPath, err)
		}
		log.Infof("Disk %s backed up to %s", disk, destFile)
		if err := removeSnapshot(snapName); err != nil {
			return err
		}
	}

	var srcConfig string
	if vm.IsVM {
		srcConfig = filepath.Join("/etc/pve/qemu-server", fmt.Sprintf("%d.conf", vm.ID))
	} else {
		srcConfig = filepath.Join("/etc/pve/lxc", fmt.Sprintf("%d.conf", vm.ID))
	}

	destConfig := filepath.Join(vmDir, "config")
	if _, err := os.Stat(srcConfig); err == nil {
		cmd := exec.Command("cp", srcConfig, destConfig)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to copy config: %v\n%s", err, string(out))
		}
		log.Infof("Config copied to %s", destConfig)
	} else {
		log.Warnf("Config file %s not found, skipping", srcConfig)
	}
	return nil
}

func removeSnapshot(snapName string) error {
	cmd := exec.Command("lvremove", "-f", "/dev/pve/"+snapName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove snapshot %s: %v\n%s", snapName, err, string(out))
	}
	log.Infof("Snapshot %s removed", snapName)
	return nil
}

func isResticConfigCorrect() bool {
	err := util.RunResticCommand(false, "snapshots")
	return err == nil
}

func isSpaceSufficient() (bool, error) {
	lvMap, err := getLVMap()
	if err != nil {
		return false, fmt.Errorf("failed to get lv map: %v", err)
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(config.Get().Backup.Dir, &stat); err != nil {
		return false, fmt.Errorf("failed to get filesystem stats: %v", err)
	}
	available := stat.Bavail * uint64(stat.Bsize)
	var totalRequired uint64
	for _, vm := range config.Get().VMs {
		for _, disk := range vm.Disks {
			parts := strings.Split(disk, "/")
			lvName := parts[len(parts)-1]
			size, ok := lvMap[lvName]
			if !ok {
				log.Warnf("disk %s not found in LVM map, skipping", disk)
				continue
			}
			totalRequired += size
		}
	}
	log.Infof("Available space: %.2f GiB", float64(available)/(1024*1024*1024))
	log.Infof("Required space for all VMs: %.2f GiB", float64(totalRequired)/(1024*1024*1024))
	if available < totalRequired {
		log.Infof("insufficient space: need %.2f GiB, but only %.2f GiB available",
			float64(totalRequired)/(1024*1024*1024),
			float64(available)/(1024*1024*1024),
		)
		return false, nil
	}
	return true, nil
}

func getLVMap() (map[string]uint64, error) {
	cmd := exec.Command("lvs", "--noheadings", "--units", "b", "--nosuffix", "-o", "lv_name,lv_size")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run lvs: %v\n%s", err, string(out))
	}

	lvMap := make(map[string]uint64)
	lines := bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		fields := strings.Fields(string(line))
		if len(fields) != 2 {
			continue
		}
		size, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			log.Warnf("cannot parse size %s for LV %s: %v", fields[1], fields[0], err)
			continue
		}
		lvMap[fields[0]] = size
	}

	return lvMap, nil
}
