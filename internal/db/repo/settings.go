package repo

import (
	"errors"

	"gorm.io/gorm"

	"prostic/internal/db"
	"prostic/internal/db/models"
)

func GetSettings() (*models.Setting, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	var settings models.Setting
	err = database.First(&settings, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func UpdatePassword(passwordHash string, needsPasswordChange bool) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Model(&models.Setting{}).
		Where("id = ?", 1).
		Updates(map[string]interface{}{
			"password_hash":         passwordHash,
			"needs_password_change": needsPasswordChange,
		}).Error
}

func UpdateBackupCron(expression string) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Model(&models.Setting{}).
		Where("id = ?", 1).
		Update("backup_cron", expression).Error
}
