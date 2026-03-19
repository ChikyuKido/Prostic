package snapshots

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitSnapshotsRouter(engine *gin.Engine) {
	group := engine.Group("/api/snapshots")
	group.Use(middlewares.Auth())
	group.GET("", listSnapshots)
	group.POST("/refresh", refreshSnapshots)
}
