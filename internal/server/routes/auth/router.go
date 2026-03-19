package auth

import (
	"github.com/gin-gonic/gin"

	"prostic/internal/server/middlewares"
)

func InitAuthRouter(engine *gin.Engine) {
	group := engine.Group("/api/auth")
	group.POST("/login", login)
	group.POST("/change-password", middlewares.Auth(), changePassword)
}
