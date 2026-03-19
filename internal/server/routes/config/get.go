package config

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appconfig "prostic/internal/config"
)

type vmResponse struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	IsVM       bool     `json:"isVM"`
	ConfigFile string   `json:"configFile"`
	Disks      []string `json:"disks"`
}

type configResponse struct {
	VMs []vmResponse `json:"vms"`
}

func getConfig(c *gin.Context) {
	cfg := appconfig.Get()
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not loaded"})
		return
	}

	vms := make([]vmResponse, 0, len(cfg.VMs))
	for _, vm := range cfg.VMs {
		vmType := "lxc"
		if vm.IsVM {
			vmType = "vm"
		}

		disks := append([]string(nil), vm.Disks...)
		vms = append(vms, vmResponse{
			ID:         vm.ID,
			Name:       vm.Name,
			Type:       vmType,
			IsVM:       vm.IsVM,
			ConfigFile: appconfig.ConfigFilePath(vm),
			Disks:      disks,
		})
	}

	c.JSON(http.StatusOK, configResponse{
		VMs: vms,
	})
}
