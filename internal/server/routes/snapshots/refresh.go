package snapshots

import (
	"net/http"

	"github.com/gin-gonic/gin"

	snapshotservice "prostic/internal/service/snapshots"
)

func refreshSnapshots(c *gin.Context) {
	count, err := snapshotservice.RefreshSnapshotCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to refresh snapshots"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
