package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type VM struct {
	Name  string   `yaml:"name"`
	ID    int      `yaml:"id"`
	Disks []string `yaml:"disks"`
}

type Restic struct {
	Repo     string `yaml:"repo"`
	Password string `yaml:"password"`
}
type Backup struct {
	Dir string `yaml:"dir"`
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
