package refresh

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitRefreshRouter(engine *gin.Engine) {
	group := engine.Group("/api/refresh")
	group.Use(middlewares.Auth())
	group.POST("", refreshAll)
}
