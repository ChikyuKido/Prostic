package backup

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	backupservice "prostic/internal/service/backups"
)

type updateSettingsRequest struct {
	CronExpression string `json:"cronExpression"`
}

func updateSettings(c *gin.Context) {
	var request updateSettingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := backupservice.UpdateCron(strings.TrimSpace(request.CronExpression)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update backup settings"})
		return
	}

	c.Status(http.StatusNoContent)
}
