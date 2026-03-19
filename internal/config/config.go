package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type VM struct {
	Name  string   `yaml:"name"`
	ID    int      `yaml:"id"`
	IsVM  bool     `yaml:"is_vm"`
	Disks []string `yaml:"disks"`
}

type Restic struct {
	EnvVars map[string]string `yaml:",inline"`
}
type Backup struct {
}
type Config struct {
	VMs    []VM   `yaml:"vms"`
	Restic Restic `yaml:"restic"`
	Backup Backup `yaml:"backup"`
}

var cfg *Config

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return err
	}

	cfg = &c
	return nil
}
func Get() *Config {
	return cfg
}

func FindVM(id int) *VM {
	if cfg == nil {
		return nil
	}

	for i := range cfg.VMs {
		if cfg.VMs[i].ID == id {
			return &cfg.VMs[i]
		}
	}

	return nil
}

func ConfigFilePath(vm VM) string {
	if vm.IsVM {
		return filepath.Join("/etc/pve/qemu-server", strconv.Itoa(vm.ID)+".conf")
	}

	return filepath.Join("/etc/pve/lxc", strconv.Itoa(vm.ID)+".conf")
}

func SnapshotExists(snapshotType string, vmID *int, srcFile string) bool {
	if vmID == nil {
		return false
	}

	vm := FindVM(*vmID)
	if vm == nil {
		return false
	}

	switch snapshotType {
	case "disk":
		for _, disk := range vm.Disks {
			if disk == srcFile {
				return true
			}
		}
	case "config":
		return srcFile == ConfigFilePath(*vm)
	}

	return false
}
