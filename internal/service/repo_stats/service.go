package repostats

import (
	"time"

	"prostic/internal/db/models"
	"prostic/internal/db/repo"
	"prostic/internal/restic"
)

func RefreshRepoStatCache() error {
	stats, err := restic.GetStats()
	if err != nil {
		return err
	}

	return repo.CreateRepoStat(models.RepoStat{
		TotalSize:              stats.TotalSize,
		TotalUncompressedSize:  stats.TotalUncompressedSize,
		CompressionRatio:       stats.CompressionRatio,
		CompressionProgress:    stats.CompressionProgress,
		CompressionSpaceSaving: stats.CompressionSpaceSaving,
		TotalBlobCount:         stats.TotalBlobCount,
		SnapshotsCount:         stats.SnapshotsCount,
		LastRefreshedAt:        time.Now(),
	})
}
