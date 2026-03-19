package backup

import (
	"net/http"

	"github.com/gin-gonic/gin"

	backupservice "prostic/internal/service/backups"
)

func listRuns(c *gin.Context) {
	runs, err := backupservice.ListRuns(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load backup runs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"runs": runs})
}
