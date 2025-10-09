package backup

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"prostic/internal/config"
	"prostic/internal/util"
	"syscall"
)

var log = util.GroupLogger("BACKUP")

func RunBackup() error {
	if config.Get() == nil {
		return errors.New("no config provided")
	}
	log.Infof("Check restic config")
	if !isResticConfigCorrect() {
		return errors.New("restic config invalid. Maybe you need to init the repo. Restic init")
	}
	log.Infof("Check backup dir")
	if _, err := os.Stat(config.Get().Backup.Dir); err != nil {
		log.Infof("Backup dir does not exist. Create it")
		err := os.MkdirAll(config.Get().Backup.Dir, 0755)
		if err != nil {
			return errors.New("backup dir could not be created")
		}
		enough, err := isSpaceSufficient()
		if err != nil {
			return fmt.Errorf("could not check space sufficient: %v", err)
		}
		if !enough {
			return errors.New("backup dir does not have enough space")
		}
	}

	return nil
}

func isResticConfigCorrect() bool {
	cmd := exec.Command("restic",
		"-r", config.Get().Restic.Repo,
		"--password-file", config.Get().Restic.Password,
		"snapshots",
	)
	_, err := cmd.CombinedOutput()
	return err == nil
}

func isSpaceSufficient() (bool, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(config.Get().Backup.Dir, &stat); err != nil {
		return false, fmt.Errorf("failed to get filesystem stats: %v", err)
	}
	available := stat.Bavail * uint64(stat.Bsize)
	var totalRequired uint64
	for _, vm := range config.Get().VMs {
		for _, disk := range vm.Disks {
			fi, err := os.Stat(disk)
			if err != nil {
				log.Warnf("cannot stat disk %s: %v. Using logical LV size.", disk, err)
				continue
			}
			totalRequired += uint64(fi.Size())
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
