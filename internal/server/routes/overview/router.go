package overview

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitOverviewRouter(engine *gin.Engine) {
	group := engine.Group("/api/overview")
	group.Use(middlewares.Auth())
	group.GET("", getOverview)
}
