package snapshots

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appconfig "prostic/internal/config"
	"prostic/internal/db/models"
	"prostic/internal/db/repo"
)

type snapshotResponse struct {
	models.Snapshot
	ExistsInConfig bool `json:"existsInConfig"`
}

func listSnapshots(c *gin.Context) {
	snapshots, err := repo.ListSnapshots()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load snapshots"})
		return
	}

	response := make([]snapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		response = append(response, snapshotResponse{
			Snapshot:       snapshot,
			ExistsInConfig: appconfig.SnapshotExists(snapshot.SnapshotType, snapshot.VMID, snapshot.SrcFile),
		})
	}

	c.JSON(http.StatusOK, gin.H{"snapshots": response})
}
