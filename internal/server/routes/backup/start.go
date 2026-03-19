package backup

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	backupservice "prostic/internal/service/backups"
	runnerservice "prostic/internal/service/runner"
)

func startBackup(c *gin.Context) {
	run, err := backupservice.StartBackup("manual")
	if err != nil {
		if errors.Is(err, runnerservice.ErrBusy) {
			c.JSON(http.StatusConflict, gin.H{"error": "another job is already running"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start backup"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"run": run})
}
