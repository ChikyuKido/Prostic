package backup

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"prostic/internal/config"
	"prostic/internal/restic"
	"strconv"
	"strings"
	"time"
)

const charset = "0123456789abcdefghijklmnopqrstuvwxyz"

func RunBackup() error {
	if config.Get() == nil {
		return errors.New("no config provided")
	}
	if !isResticConfigCorrect() {
		return errors.New("restic config invalid. Maybe you need to init the repo. Run restic init")
	}

	backupID := randomID(10)
	for _, vm := range config.Get().VMs {
		if err := runVMBackup(vm, backupID); err != nil {
			return fmt.Errorf("could not backup vm %d: %v", vm.ID, err)
		}
	}

	fmt.Println(cGreen + "Backup completed successfully" + cReset)
	return nil
}

func runVMBackup(vm config.VM, backupID string) error {
	vmPrefix := "lxc"
	if vm.IsVM {
		vmPrefix = "vm"
	}
	today := time.Now().Format("2006-01-02")

	fmt.Printf(cBlue+"Backing up %s (id=%d)"+cReset+"\n", vm.Name, vm.ID)

	for _, disk := range vm.Disks {
		parts := strings.Split(disk, "/")
		lvName := parts[len(parts)-1]
		snapName := lvName + "-snap"
		snapPath := filepath.Join(filepath.Dir(disk), snapName)
		destFile := filepath.Join(fmt.Sprintf("%s-%d", vmPrefix, vm.ID), parts[2], lvName+".raw")

		if _, err := os.Stat(disk + "-snap"); err == nil {
			fmt.Println(cCyan + "Removing old snapshot " + cReset + disk + "-snap")
			if err := removeSnapshot(snapPath); err != nil {
				return err
			}
		}

		cmd := exec.Command("/usr/sbin/lvcreate", "-s", "-n", snapName, "-L", "5G", disk)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to create snapshot for %s: %v\n%s", disk, err, string(out))
		}

		totalGB := probeLVSizeGB(disk)
		if totalGB <= 0 {
			totalGB = 5.0
		}

		fmt.Printf(cCyan+"Backing up %s (%.2f GiB) -> %s"+cReset+"\n", disk, totalGB, destFile)

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

		err = restic.RunResticJSONStream(args, func(obj map[string]interface{}) { handleResticMsg(obj, totalGB) })
		_ = removeSnapshot(snapPath)
		if err != nil {
			return fmt.Errorf("restic backup failed for %s: %v", snapPath, err)
		}
		fmt.Println()
		fmt.Println(cGreen + "Finished " + cReset + disk)
	}

	var srcConfig string
	if vm.IsVM {
		srcConfig = filepath.Join("/etc/pve/qemu-server", fmt.Sprintf("%d.conf", vm.ID))
	} else {
		srcConfig = filepath.Join("/etc/pve/lxc", fmt.Sprintf("%d.conf", vm.ID))
	}
	destConfig := filepath.Join(fmt.Sprintf("%s-%d", vmPrefix, vm.ID), "config")

	if _, err := os.Stat(srcConfig); err == nil {
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

		err := restic.RunResticJSONStream(args, func(obj map[string]interface{}) { handleResticMsg(obj, 1.0/1024/1024) })
		fmt.Println()
		if err != nil {
			return fmt.Errorf("restic backup failed for config %s: %v", srcConfig, err)
		}
	} else {
		fmt.Println(cCyan + "Config file not found, skipping:" + cReset + " " + srcConfig)
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

func handleResticMsg(obj map[string]interface{}, totalGB float64) {
	mt, _ := obj["message_type"].(string)
	switch mt {
	case "status", "progress", "snapshot":
		var done float64
		if v, ok := obj["bytes_done"]; ok {
			switch vv := v.(type) {
			case float64:
				done = vv
			case int64:
				done = float64(vv)
			case int:
				done = float64(vv)
			}
		}
		doneGB := done / (1024.0 * 1024.0 * 1024.0)
		fmt.Printf("\r%s", renderBar(doneGB, totalGB))
	case "error":
		fmt.Println()
		fmt.Println(cBold+"RESTIC ERROR:"+cReset, obj["message"])
	default:
		if m, ok := obj["message"].(string); ok && m != "" {
			fmt.Println(m)
		}
	}
}

func renderBar(currentGB, totalGB float64) string {
	width := 40
	var perc float64
	if totalGB > 0 {
		perc = currentGB / totalGB
		if perc < 0 {
			perc = 0
		}
		if perc > 1 {
			perc = 1
		}
	} else {
		perc = 0
	}
	filled := int(perc * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	if totalGB > 0 {
		return fmt.Sprintf("[%s] %.2f / %.2f GiB", bar, currentGB, totalGB)
	}
	return fmt.Sprintf("[%s] %.2f GiB", bar, currentGB)
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
