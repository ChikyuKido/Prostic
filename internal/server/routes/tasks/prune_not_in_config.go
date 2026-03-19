package tasks

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/models"
	pruneservice "prostic/internal/service/prune"
	taskservice "prostic/internal/service/tasks"
)

type pruneRequest struct {
	Snapshots []pruneservice.SnapshotCandidate `json:"snapshots"`
}

func pruneNotInConfig(c *gin.Context) {
	if c.Query("confirm") != "true" {
		snapshots, err := pruneservice.PreviewNotInConfig()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare prune task"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"snapshots": snapshots})
		return
	}

	var request pruneRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	task, err := taskservice.StartBackgroundTask(pruneservice.TaskPurposePruneNotInConfig, func(_ *models.Task) (string, error) {
		return pruneservice.RunNotInConfig(request.Snapshots)
	})
	if err != nil {
		if errors.Is(err, taskservice.ErrTaskRunning) {
			c.JSON(http.StatusConflict, gin.H{"error": "another task is already running"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run prune task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"task":         task,
		"deletedCount": len(request.Snapshots),
	})
}
