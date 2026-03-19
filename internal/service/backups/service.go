package backups

import (
	"errors"
	"fmt"
	"strings"
	"sync"
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

type LiveStatus struct {
	Running            bool       `json:"running"`
	RunnerBusy         bool       `json:"runnerBusy"`
	RunnerKind         string     `json:"runnerKind,omitempty"`
	RunnerPurpose      string     `json:"runnerPurpose,omitempty"`
	BackupRunID        *uint      `json:"backupRunID,omitempty"`
	BackupID           string     `json:"backupID,omitempty"`
	Trigger            string     `json:"trigger,omitempty"`
	StartedAt          *time.Time `json:"startedAt,omitempty"`
	TotalItems         int        `json:"totalItems"`
	CompletedItems     int        `json:"completedItems"`
	CurrentVMID        *int       `json:"currentVMID,omitempty"`
	CurrentVMName      string     `json:"currentVMName,omitempty"`
	CurrentItemType    string     `json:"currentItemType,omitempty"`
	CurrentSrcFile     string     `json:"currentSrcFile,omitempty"`
	CurrentDestFile    string     `json:"currentDestFile,omitempty"`
	CurrentBytesDone   int64      `json:"currentBytesDone"`
	CurrentBytesTotal  int64      `json:"currentBytesTotal"`
	CurrentItemStarted *time.Time `json:"currentItemStarted,omitempty"`
	LastMessage        string     `json:"lastMessage,omitempty"`
	CronExpression     string     `json:"cronExpression"`
}

var (
	liveMu      sync.Mutex
	liveStatus  = LiveStatus{}
	schedulerMu sync.Mutex
	lastTickKey string
)

func StartBackup(trigger string) (*models.BackupRun, error) {
	handle, err := runnerservice.Start("backup", "backup")
	if err != nil {
		return nil, err
	}

	run := &models.BackupRun{
		Trigger:    trigger,
		Status:     StatusRunning,
		StartedAt:  time.Now(),
		TotalItems: 0,
	}
	if err := repo.CreateBackupRun(run); err != nil {
		handle.Release()
		return nil, err
	}

	setLiveStatus(func(status *LiveStatus) {
		status.Running = true
		status.BackupRunID = &run.ID
		status.Trigger = trigger
		status.StartedAt = &run.StartedAt
		status.TotalItems = 0
		status.CompletedItems = 0
		status.CurrentVMID = nil
		status.CurrentVMName = ""
		status.CurrentItemType = ""
		status.CurrentSrcFile = ""
		status.CurrentDestFile = ""
		status.CurrentBytesDone = 0
		status.CurrentBytesTotal = 0
		status.CurrentItemStarted = nil
		status.LastMessage = ""
		status.BackupID = ""
	})

	go func(run *models.BackupRun) {
		defer handle.Release()

		var logs strings.Builder
		observer := ObserverFunc(func(event Event) {
			switch event.Type {
			case EventRunStarted:
				setLiveStatus(func(status *LiveStatus) {
					status.BackupID = event.BackupID
					status.TotalItems = event.TotalItems
				})
				_ = repo.UpdateBackupRun(run.ID, map[string]interface{}{
					"backup_id":   event.BackupID,
					"total_items": event.TotalItems,
				})
			case EventItemStarted:
				now := time.Now()
				setLiveStatus(func(status *LiveStatus) {
					status.CompletedItems = event.CompletedItems
					status.TotalItems = event.TotalItems
					status.CurrentBytesDone = 0
					status.CurrentBytesTotal = event.BytesTotal
					status.CurrentItemStarted = &now
					if event.Item != nil {
						status.CurrentVMName = event.Item.VM.Name
						status.CurrentItemType = event.Item.ItemType
						status.CurrentSrcFile = event.Item.SrcFile
						status.CurrentDestFile = event.Item.DestFile
						status.CurrentVMID = &event.Item.VM.ID
					}
				})
			case EventItemProgress:
				setLiveStatus(func(status *LiveStatus) {
					status.CompletedItems = event.CompletedItems
					status.TotalItems = event.TotalItems
					status.CurrentBytesDone = event.BytesDone
					status.CurrentBytesTotal = event.BytesTotal
				})
			case EventItemDone:
				setLiveStatus(func(status *LiveStatus) {
					status.CompletedItems = event.CompletedItems
					status.TotalItems = event.TotalItems
					status.CurrentBytesDone = event.BytesDone
					status.CurrentBytesTotal = event.BytesTotal
				})
				_ = repo.UpdateBackupRun(run.ID, map[string]interface{}{
					"completed_items": event.CompletedItems,
					"total_items":     event.TotalItems,
				})
			case EventLog:
				if event.Message != "" {
					if logs.Len() > 0 {
						logs.WriteString("\n")
					}
					logs.WriteString(event.Message)
					setLiveStatus(func(status *LiveStatus) {
						status.LastMessage = event.Message
					})
				}
			case EventRunFailed:
				if event.Message != "" {
					if logs.Len() > 0 {
						logs.WriteString("\n")
					}
					logs.WriteString(event.Message)
				}
			}
		})

		err := RunBackupWithObserver(observer)
		finalLogs := logs.String()
		if err != nil {
			if finalLogs != "" && !strings.HasSuffix(finalLogs, "\n") {
				finalLogs += "\n"
			}
			finalLogs += fmt.Sprintf("Error: %v", err)
			_ = repo.FinishBackupRun(run.ID, StatusFailed, finalLogs, getLiveStatus().BackupID, getLiveStatus().CompletedItems)
		} else {
			_ = repo.FinishBackupRun(run.ID, StatusSuccess, finalLogs, getLiveStatus().BackupID, getLiveStatus().CompletedItems)
		}

		clearLiveStatus()
	}(run)

	return run, nil
}

func GetLiveStatus() LiveStatus {
	status := getLiveStatus()
	settings, err := repo.GetSettings()
	if err == nil && settings != nil {
		status.CronExpression = settings.BackupCron
	}

	runner := runnerservice.GetStatus()
	status.RunnerBusy = runner.Running
	status.RunnerKind = runner.Kind
	status.RunnerPurpose = runner.Purpose
	if !status.Running && runner.Kind == "backup" {
		status.Running = true
		status.StartedAt = runner.StartedAt
	}

	return status
}

func UpdateCron(expression string) error {
	expression = strings.TrimSpace(expression)
	if expression != "" {
		if _, err := cronMatches(expression, time.Now().In(time.Local)); err != nil {
			return errors.New("invalid cron expression")
		}
	}

	return repo.UpdateBackupCron(expression)
}

func ListRuns(limit int) ([]models.BackupRun, error) {
	return repo.ListBackupRuns(limit)
}

func SchedulerTick(now time.Time) {
	settings, err := repo.GetSettings()
	if err != nil || settings == nil || strings.TrimSpace(settings.BackupCron) == "" {
		return
	}

	matches, err := cronMatches(settings.BackupCron, now)
	if err != nil || !matches {
		return
	}

	key := now.In(time.Local).Format("2006-01-02 15:04")
	schedulerMu.Lock()
	if lastTickKey == key {
		schedulerMu.Unlock()
		return
	}
	lastTickKey = key
	schedulerMu.Unlock()

	_, _ = StartBackup("scheduled")
}

func setLiveStatus(update func(*LiveStatus)) {
	liveMu.Lock()
	defer liveMu.Unlock()
	update(&liveStatus)
}

func clearLiveStatus() {
	liveMu.Lock()
	defer liveMu.Unlock()
	liveStatus = LiveStatus{}
}

func getLiveStatus() LiveStatus {
	liveMu.Lock()
	defer liveMu.Unlock()
	return liveStatus
}
