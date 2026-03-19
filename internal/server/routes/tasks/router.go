package tasks

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitTasksRouter(engine *gin.Engine) {
	group := engine.Group("/api/tasks")
	group.Use(middlewares.Auth())
	group.GET("", listTasks)
	group.GET("/status", getStatus)
	group.POST("/delete-snapshot", deleteSnapshot)
	group.POST("/delete-backup-id", deleteBackupID)
	group.POST("/prune-not-in-config", pruneNotInConfig)
}
