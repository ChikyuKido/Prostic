package overview

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/repo"
)

func getOverview(c *gin.Context) {
	overview, err := repo.GetSnapshotOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load overview"})
		return
	}

	c.JSON(http.StatusOK, overview)
}
