package tasks

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/models"
	pruneservice "prostic/internal/service/prune"
	taskservice "prostic/internal/service/tasks"
)

type deleteSnapshotRequest struct {
	Snapshot pruneservice.SnapshotCandidate `json:"snapshot"`
}

func deleteSnapshot(c *gin.Context) {
	var request deleteSnapshotRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if request.Snapshot.SnapshotID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "snapshotID is required"})
		return
	}

	task, err := taskservice.StartBackgroundTask(pruneservice.TaskPurposeDeleteSnapshot, func(_ *models.Task) (string, error) {
		return pruneservice.RunDeleteSnapshot(request.Snapshot)
	})
	if err != nil {
		if errors.Is(err, taskservice.ErrTaskRunning) {
			c.JSON(http.StatusConflict, gin.H{"error": "another task is already running"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to run delete snapshot task"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"task": task})
}
