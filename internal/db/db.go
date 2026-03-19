package db

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prostic/internal/db/models"
	"prostic/internal/util"
)

var (
	instance      *gorm.DB
	initErr       error
	once          sync.Once
	defaultDBPath = filepath.Join("data", "prostic.db")
)

func Get() (*gorm.DB, error) {
	once.Do(func() {
		dbPath := os.Getenv("PROSTIC_DB_PATH")
		if dbPath == "" {
			dbPath = defaultDBPath
		}

		dir := filepath.Dir(dbPath)
		if dir != "." {
			initErr = os.MkdirAll(dir, 0o755)
			if initErr != nil {
				return
			}
		}

		instance, initErr = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if initErr != nil {
			return
		}

		initErr = instance.AutoMigrate(&models.Setting{}, &models.Snapshot{}, &models.RepoStat{}, &models.Task{}, &models.BackupRun{})
		if initErr != nil {
			return
		}

		initErr = ensureDefaultSettings()
	})

	return instance, initErr
}

func ensureDefaultSettings() error {
	var settings models.Setting
	err := instance.First(&settings, 1).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	passwordHash, err := util.HashPassword("admin")
	if err != nil {
		return err
	}

	settings = models.Setting{
		ID:                  1,
		PasswordHash:        passwordHash,
		NeedsPasswordChange: true,
		BackupCron:          "",
	}

	return instance.Create(&settings).Error
}
