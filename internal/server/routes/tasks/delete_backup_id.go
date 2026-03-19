package tasks

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/models"
	pruneservice "prostic/internal/service/prune"
	taskservice "prostic/internal/service/tasks"
)

type deleteBackupIDRequest struct {
	BackupID  string                           `json:"backupID"`
	Snapshots []pruneservice.SnapshotCandidate `json:"snapshots"`
}

func deleteBackupID(c *gin.Context) {
	var request deleteBackupIDRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if request.BackupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "backupID is required"})
		return
	}

	if c.Query("confirm") != "true" {
		snapshots, err := pruneservice.PreviewBackupID(request.BackupID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare delete backup task"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"snapshots": snapshots})
		return
	}

	task, err := taskservice.StartBackgroundTask(pruneservice.TaskPurposeDeleteBackupID, func(_ *models.Task) (string, error) {
		return pruneservice.RunDeleteBackupID(request.Snapshots)
	})
	if err != nil {
		if errors.Is(err, taskservice.ErrTaskRunning) {
			c.JSON(http.StatusConflict, gin.H{"error": "another task is already running"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run delete backup task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"task":         task,
		"deletedCount": len(request.Snapshots),
	})
}
