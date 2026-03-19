package repo

import (
	"time"

	"gorm.io/gorm"

	"prostic/internal/db"
	"prostic/internal/db/models"
)

type SnapshotOverview struct {
	TotalSnapshots int64      `json:"totalSnapshots"`
	TotalBackups   int64      `json:"totalBackups"`
	TotalVMs       int64      `json:"totalVMs"`
	DiskSnapshots  int64      `json:"diskSnapshots"`
	Configs        int64      `json:"configSnapshots"`
	LatestSnapshot *time.Time `json:"latestSnapshot"`
}

func ReplaceSnapshots(snapshots []models.Snapshot) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM snapshots").Error; err != nil {
			return err
		}

		if len(snapshots) == 0 {
			return nil
		}

		return tx.Create(&snapshots).Error
	})
}

func ListSnapshots() ([]models.Snapshot, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	var snapshots []models.Snapshot
	if err := database.Order("time desc").Find(&snapshots).Error; err != nil {
		return nil, err
	}

	return snapshots, nil
}

func GetSnapshotOverview() (*SnapshotOverview, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	overview := &SnapshotOverview{}
	if err := database.Model(&models.Snapshot{}).Count(&overview.TotalSnapshots).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&models.Snapshot{}).Distinct("backup_id").Where("backup_id <> ''").Count(&overview.TotalBackups).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&models.Snapshot{}).Distinct("vm_id").Where("vm_id IS NOT NULL").Count(&overview.TotalVMs).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&models.Snapshot{}).Where("snapshot_type = ?", "disk").Count(&overview.DiskSnapshots).Error; err != nil {
		return nil, err
	}
	if err := database.Model(&models.Snapshot{}).Where("snapshot_type = ?", "config").Count(&overview.Configs).Error; err != nil {
		return nil, err
	}

	var latest models.Snapshot
	if err := database.Order("time desc").First(&latest).Error; err == nil {
		overview.LatestSnapshot = &latest.Time
	}

	return overview, nil
}
