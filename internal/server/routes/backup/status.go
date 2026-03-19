package backup

import (
	"net/http"

	"github.com/gin-gonic/gin"

	backupservice "prostic/internal/service/backups"
)

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, backupservice.GetLiveStatus())
}
