package backup

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitBackupRouter(engine *gin.Engine) {
	group := engine.Group("/api/backup")
	group.Use(middlewares.Auth())
	group.GET("/status", getStatus)
	group.GET("/runs", listRuns)
	group.POST("/start", startBackup)
	group.PUT("/settings", updateSettings)
}
