package snapshots

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/repo"
)

func listSnapshots(c *gin.Context) {
	snapshots, err := repo.ListSnapshots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"snapshots": snapshots})
}
