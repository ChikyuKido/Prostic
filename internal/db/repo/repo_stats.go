package repo

import (
	"errors"

	"gorm.io/gorm"
	"prostic/internal/db"
	"prostic/internal/db/models"
)

func CreateRepoStat(repoStat models.RepoStat) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Create(&repoStat).Error
}

func GetLatestRepoStat() (*models.RepoStat, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	var repoStat models.RepoStat
	if err := database.Order("last_refreshed_at desc, id desc").First(&repoStat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &repoStat, nil
}

func ListRepoStats(limit int) ([]models.RepoStat, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	query := database.Model(&models.RepoStat{})
	if limit > 0 {
		query = query.Order("last_refreshed_at desc, id desc").Limit(limit)
	}

	var repoStats []models.RepoStat
	if err := query.Find(&repoStats).Error; err != nil {
		return nil, err
	}

	for i, j := 0, len(repoStats)-1; i < j; i, j = i+1, j-1 {
		repoStats[i], repoStats[j] = repoStats[j], repoStats[i]
	}

	return repoStats, nil
}
