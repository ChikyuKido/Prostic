package repo

import (
	"time"

	"prostic/internal/db"
	"prostic/internal/db/models"
)

func CreateTask(task *models.Task) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	return database.Create(task).Error
}

func UpdateTask(taskID uint, status string, logs string, finishedAt *time.Time) error {
	database, err := db.Get()
	if err != nil {
		return err
	}

	updates := map[string]interface{}{
		"status":      status,
		"logs":        logs,
		"finished_at": finishedAt,
	}

	return database.Model(&models.Task{}).Where("id = ?", taskID).Updates(updates).Error
}

func ListTasks() ([]models.Task, error) {
	database, err := db.Get()
	if err != nil {
		return nil, err
	}

	var tasks []models.Task
	if err := database.Order("started_at desc, id desc").Find(&tasks).Error; err != nil {
		return nil, err
	}

	return tasks, nil
}
