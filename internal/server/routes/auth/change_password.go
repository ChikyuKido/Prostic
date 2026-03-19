package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/repo"
	"prostic/internal/util"
)

type changePasswordRequest struct {
	Password string `json:"password"`
}

func changePassword(c *gin.Context) {
	var request changePasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if strings.TrimSpace(request.Password) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password is required"})
		return
	}

	passwordHash, err := util.HashPassword(request.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := repo.UpdatePassword(passwordHash, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.Status(http.StatusNoContent)
}
