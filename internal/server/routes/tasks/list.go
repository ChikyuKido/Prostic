package tasks

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/repo"
)

func listTasks(c *gin.Context) {
	tasks, err := repo.ListTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}
