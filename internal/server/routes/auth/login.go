package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/repo"
	"prostic/internal/util"
)

type loginRequest struct {
	Password string `json:"password"`
}

func login(c *gin.Context) {
	var request loginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(request.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	settings, err := repo.GetSettings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load settings"})
		return
	}
	if settings == nil || settings.PasswordHash == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "password not configured"})
		return
	}

	if err := util.CheckPassword(settings.PasswordHash, request.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token, err := util.CreateJWT()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":               token,
		"needsPasswordChange": settings.NeedsPasswordChange,
	})
}
