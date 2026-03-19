package repo

import (
	"time"

	"prostic/internal/db"
	"prostic/internal/db/models"
)

func CreateBackupRun(run *models.BackupRun) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Create(run).Error
}

func UpdateBackupRun(runID uint, updates map[string]interface{}) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Model(&models.BackupRun{}).Where("id = ?", runID).Updates(updates).Error
}

func FinishBackupRun(runID uint, status string, logs string, backupID string, completedItems int) error {
	finishedAt := time.Now()
	return UpdateBackupRun(runID, map[string]interface{}{
		"status":          status,
		"logs":            logs,
		"backup_id":       backupID,
		"completed_items": completedItems,
		"finished_at":     &finishedAt,
	})
}

func ListBackupRuns(limit int) ([]models.BackupRun, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	query := database.Order("started_at desc, id desc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var runs []models.BackupRun
	if err := query.Find(&runs).Error; err != nil {
		return nil, err
	}

	return runs, nil
}
