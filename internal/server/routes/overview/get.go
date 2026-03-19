package overview

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"prostic/internal/db/repo"
)

type overviewHistoryPoint struct {
	Timestamp             time.Time `json:"timestamp"`
	TotalSize             int64     `json:"totalSize"`
	TotalUncompressedSize int64     `json:"totalUncompressedSize"`
}

type overviewResponse struct {
	TotalSnapshots         int64                  `json:"totalSnapshots"`
	TotalBackups           int64                  `json:"totalBackups"`
	TotalVMs               int64                  `json:"totalVMs"`
	DiskSnapshots          int64                  `json:"diskSnapshots"`
	ConfigSnapshots        int64                  `json:"configSnapshots"`
	LatestSnapshot         *time.Time             `json:"latestSnapshot"`
	TotalSize              int64                  `json:"totalSize"`
	TotalUncompressedSize  int64                  `json:"totalUncompressedSize"`
	CompressionRatio       float64                `json:"compressionRatio"`
	CompressionSpaceSaving float64                `json:"compressionSpaceSaving"`
	TotalBlobCount         int64                  `json:"totalBlobCount"`
	RepoSnapshotsCount     int64                  `json:"repoSnapshotsCount"`
	LastRefreshedAt        *time.Time             `json:"lastRefreshedAt"`
	History                []overviewHistoryPoint `json:"history"`
}

func getOverview(c *gin.Context) {
	overview, err := repo.GetSnapshotOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load overview"})
		return
	}

	repoStat, err := repo.GetLatestRepoStat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load overview"})
		return
	}

	history, err := repo.ListRepoStats(60)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load overview"})
		return
	}

	response := overviewResponse{
		TotalSnapshots:  overview.TotalSnapshots,
		TotalBackups:    overview.TotalBackups,
		TotalVMs:        overview.TotalVMs,
		DiskSnapshots:   overview.DiskSnapshots,
		ConfigSnapshots: overview.Configs,
		LatestSnapshot:  overview.LatestSnapshot,
	}

	if repoStat != nil {
		response.TotalSize = repoStat.TotalSize
		response.TotalUncompressedSize = repoStat.TotalUncompressedSize
		response.CompressionRatio = repoStat.CompressionRatio
		response.CompressionSpaceSaving = repoStat.CompressionSpaceSaving
		response.TotalBlobCount = repoStat.TotalBlobCount
		response.RepoSnapshotsCount = repoStat.SnapshotsCount
		response.LastRefreshedAt = &repoStat.LastRefreshedAt
	}

	response.History = make([]overviewHistoryPoint, 0, len(history))
	for _, point := range history {
		response.History = append(response.History, overviewHistoryPoint{
			Timestamp:             point.LastRefreshedAt,
			TotalSize:             point.TotalSize,
			TotalUncompressedSize: point.TotalUncompressedSize,
		})
	}

	c.JSON(http.StatusOK, response)
}
