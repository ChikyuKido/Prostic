package tasks

import (
	"net/http"

	"github.com/gin-gonic/gin"

	taskservice "prostic/internal/service/tasks"
)

func getStatus(c *gin.Context) {
	c.JSON(http.StatusOK, taskservice.GetStatus())
}
