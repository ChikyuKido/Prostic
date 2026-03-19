package tasks

import (
	"fmt"
	"strings"
	"time"

	"prostic/internal/db/models"
	"prostic/internal/db/repo"
	runnerservice "prostic/internal/service/runner"
)

const (
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

var ErrTaskRunning = runnerservice.ErrBusy

type StatusResponse struct {
	Running   bool       `json:"running"`
	Kind      string     `json:"kind,omitempty"`
	Purpose   string     `json:"purpose,omitempty"`
	StartedAt *time.Time `json:"startedAt,omitempty"`
}

func StartTask(purpose string) (*models.Task, error) {
	task := &models.Task{
		Purpose:   purpose,
		Status:    StatusRunning,
		StartedAt: time.Now(),
	}

	if err := repo.CreateTask(task); err != nil {
		return nil, err
	}

	return task, nil
}

func FinishTask(taskID uint, status string, logs string) error {
	finishedAt := time.Now()
	return repo.UpdateTask(taskID, status, logs, &finishedAt)
}

func StartBackgroundTask(purpose string, run func(task *models.Task) (string, error)) (*models.Task, error) {
	handle, err := runnerservice.Start("task", purpose)
	if err != nil {
		if err == runnerservice.ErrBusy {
			return nil, ErrTaskRunning
		}
		return nil, err
	}

	task, err := StartTask(purpose)
	if err != nil {
		handle.Release()
		return nil, err
	}

	go func(task *models.Task) {
		defer handle.Release()

		logs, runErr := run(task)
		status := StatusSuccess
		if runErr != nil {
			status = StatusFailed
			if !strings.Contains(logs, runErr.Error()) {
				if logs != "" && !strings.HasSuffix(logs, "\n") {
					logs += "\n"
				}
				logs += fmt.Sprintf("Error: %v\n", runErr)
			}
		}

		_ = FinishTask(task.ID, status, logs)
	}(task)

	return task, nil
}

func GetStatus() StatusResponse {
	status := runnerservice.GetStatus()
	return StatusResponse{
		Running:   status.Running,
		Kind:      status.Kind,
		Purpose:   status.Purpose,
		StartedAt: status.StartedAt,
	}
}
