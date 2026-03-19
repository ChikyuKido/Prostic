package backups

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"prostic/internal/config"
	"prostic/internal/restic"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyz"

func RunBackup() error {
	return RunBackupWithObserver(consoleObserver{})
}

func RunBackupWithObserver(observer Observer) error {
	if config.Get() == nil {
		return errors.New("no config provided")
	}
	if !isResticConfigCorrect() {
		return errors.New("restic config invalid. Maybe you need to init the repo. Run restic init")
	}

	observer = normalizeObserver(observer)
	backupID := randomID(10)
	totalItems := countBackupItems()
	completedItems := 0
	observer.OnEvent(Event{
		Type:       EventRunStarted,
		BackupID:   backupID,
		TotalItems: totalItems,
	})

	for _, vm := range config.Get().VMs {
		if err := runVMBackup(vm, backupID, observer, totalItems, &completedItems); err != nil {
			observer.OnEvent(Event{
				Type:           EventRunFailed,
				BackupID:       backupID,
				TotalItems:     totalItems,
				CompletedItems: completedItems,
				Message:        err.Error(),
			})
			return fmt.Errorf("could not backup vm %d: %v", vm.ID, err)
		}
	}

	observer.OnEvent(Event{
		Type:           EventRunDone,
		BackupID:       backupID,
		TotalItems:     totalItems,
		CompletedItems: completedItems,
	})
	return nil
}

func runVMBackup(vm config.VM, backupID string, observer Observer, totalItems int, completedItems *int) error {
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
		destFile := filepath.Join(fmt.Sprintf("%s-%d", vmPrefix, vm.ID), parts[2], lvName+".raw")

		if _, err := os.Stat(disk + "-snap"); err == nil {
			observer.OnEvent(Event{Type: EventLog, BackupID: backupID, Message: "Removing old snapshot " + disk + "-snap"})
			if err := removeSnapshot(snapPath); err != nil {
				return err
			}
		}

		isThin, err := isThinLV(disk)
		if err != nil {
			return err
		}
		var cmd *exec.Cmd
		if isThin {
			cmd = exec.Command("/usr/sbin/lvcreate", "-s", "-n", snapName, disk)
		} else {
			cmd = exec.Command("/usr/sbin/lvcreate", "-s", "-n", snapName, "-L", "5G", disk)
		}

		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create snapshot for %s: %v\n%s", disk, err, string(out))
		}

		if isThin {
			cmd = exec.Command("/usr/sbin/lvchange", "-ay", "-Ky", snapPath)
			out, err = cmd.CombinedOutput()
			if err != nil {
				_ = removeSnapshot(snapPath)
				return fmt.Errorf("failed to activate thin snapshot %s: %v\n%s", snapPath, err, string(out))
			}
		}

		totalGB := probeLVSizeGB(disk)
		if totalGB <= 0 {
			totalGB = 5.0
		}
		item := &Item{
			VM:       vm,
			ItemType: "disk",
			SrcFile:  disk,
			DestFile: destFile,
		}
		observer.OnEvent(Event{
			Type:           EventItemStarted,
			BackupID:       backupID,
			TotalItems:     totalItems,
			CompletedItems: *completedItems,
			Item:           item,
			BytesTotal:     int64(totalGB * 1024 * 1024 * 1024),
		})

		args := []string{
			"backup",
			"--stdin-from-command",
			"--json",
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
			"/bin/dd", "if=" + snapPath, "bs=4M",
		}

		err = restic.RunResticJSONStream(args, func(obj map[string]interface{}) {
			handleResticMsg(obj, func(doneBytes int64, message string) {
				if message != "" {
					observer.OnEvent(Event{Type: EventLog, BackupID: backupID, Message: message})
				}
				if doneBytes >= 0 {
					observer.OnEvent(Event{
						Type:           EventItemProgress,
						BackupID:       backupID,
						TotalItems:     totalItems,
						CompletedItems: *completedItems,
						Item:           item,
						BytesDone:      doneBytes,
						BytesTotal:     int64(totalGB * 1024 * 1024 * 1024),
					})
				}
			})
		})
		_ = removeSnapshot(snapPath)
		if err != nil {
			return fmt.Errorf("restic backup failed for %s: %v", snapPath, err)
		}
		(*completedItems)++
		observer.OnEvent(Event{
			Type:           EventItemDone,
			BackupID:       backupID,
			TotalItems:     totalItems,
			CompletedItems: *completedItems,
			Item:           item,
			BytesDone:      int64(totalGB * 1024 * 1024 * 1024),
			BytesTotal:     int64(totalGB * 1024 * 1024 * 1024),
		})
	}

	var srcConfig string
	if vm.IsVM {
		srcConfig = filepath.Join("/etc/pve/qemu-server", fmt.Sprintf("%d.conf", vm.ID))
	} else {
		srcConfig = filepath.Join("/etc/pve/lxc", fmt.Sprintf("%d.conf", vm.ID))
	}
	destConfig := filepath.Join(fmt.Sprintf("%s-%d", vmPrefix, vm.ID), "config")

	if _, err := os.Stat(srcConfig); err == nil {
		fileInfo, err := os.Stat(srcConfig)
		if err != nil {
			return err
		}
		item := &Item{
			VM:       vm,
			ItemType: "config",
			SrcFile:  srcConfig,
			DestFile: destConfig,
		}
		observer.OnEvent(Event{
			Type:           EventItemStarted,
			BackupID:       backupID,
			TotalItems:     totalItems,
			CompletedItems: *completedItems,
			Item:           item,
			BytesTotal:     fileInfo.Size(),
		})
		args := []string{
			"backup",
			"--stdin-from-command",
			"--json",
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
			"/bin/dd", "if=" + srcConfig, "bs=4M",
		}
		err = restic.RunResticJSONStream(args, func(obj map[string]interface{}) {
			handleResticMsg(obj, func(doneBytes int64, message string) {
				if message != "" {
					observer.OnEvent(Event{Type: EventLog, BackupID: backupID, Message: message})
				}
				if doneBytes >= 0 {
					observer.OnEvent(Event{
						Type:           EventItemProgress,
						BackupID:       backupID,
						TotalItems:     totalItems,
						CompletedItems: *completedItems,
						Item:           item,
						BytesDone:      doneBytes,
						BytesTotal:     fileInfo.Size(),
					})
				}
			})
		})
		if err != nil {
			return fmt.Errorf("restic backup failed for config %s: %v", srcConfig, err)
		}
		(*completedItems)++
		observer.OnEvent(Event{
			Type:           EventItemDone,
			BackupID:       backupID,
			TotalItems:     totalItems,
			CompletedItems: *completedItems,
			Item:           item,
			BytesDone:      fileInfo.Size(),
			BytesTotal:     fileInfo.Size(),
		})
	} else {
		observer.OnEvent(Event{Type: EventLog, BackupID: backupID, Message: "Config file not found, skipping: " + srcConfig})
	}

	return nil
}

func removeSnapshot(path string) error {
	cmd := exec.Command("/usr/sbin/lvremove", "-f", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove snapshot %s: %v\n%s", path, err, string(out))
	}
	return nil
}

func isResticConfigCorrect() bool {
	err := restic.RunResticCommand(false, "snapshots")
	return err == nil
}

func randomID(n int) string {
	id := make([]byte, n)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return string(id)
}

func isThinLV(lvPath string) (bool, error) {
	cmd := exec.Command("/usr/sbin/lvs", "--noheadings", "-o", "lv_attr", lvPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to inspect lv %s: %v\n%s", lvPath, err, string(out))
	}

	attr := strings.TrimSpace(string(out))
	if attr == "" {
		return false, fmt.Errorf("empty lv_attr for %s", lvPath)
	}

	return attr[0] == 'V', nil
}

func handleResticMsg(obj map[string]interface{}, report func(doneBytes int64, message string)) {
	mt, _ := obj["message_type"].(string)
	switch mt {
	case "status", "progress", "snapshot":
		var done int64
		if v, ok := obj["bytes_done"]; ok {
			switch vv := v.(type) {
			case float64:
				done = int64(vv)
			case int64:
				done = vv
			case int:
				done = int64(vv)
			}
		}
		report(done, "")
	case "error":
		if message, ok := obj["message"].(string); ok {
			report(-1, "RESTIC ERROR: "+message)
		}
	default:
		if m, ok := obj["message"].(string); ok && m != "" {
			report(-1, m)
		}
	}
}

func probeLVSizeGB(lvPath string) float64 {
	cmd := exec.Command("lvs", "--units", "g", "-o", "lv_size", "--noheadings", "--nosuffix", lvPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return -1
	}
	s := strings.TrimSpace(string(out))
	s = strings.Fields(s)[0]
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1
	}
	return v
}

func countBackupItems() int {
	if config.Get() == nil {
		return 0
	}

	total := 0
	for _, vm := range config.Get().VMs {
		total += len(vm.Disks)
		configFile := filepath.Join("/etc/pve/lxc", fmt.Sprintf("%d.conf", vm.ID))
		if vm.IsVM {
			configFile = filepath.Join("/etc/pve/qemu-server", fmt.Sprintf("%d.conf", vm.ID))
		}
		if _, err := os.Stat(configFile); err == nil {
			total++
		}
	}

	return total
}
