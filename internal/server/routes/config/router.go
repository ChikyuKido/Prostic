package config

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitConfigRouter(engine *gin.Engine) {
	group := engine.Group("/api/config")
	group.Use(middlewares.Auth())
	group.GET("", getConfig)
}
