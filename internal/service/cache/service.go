package cache

import (
	repostatsservice "prostic/internal/service/repo_stats"
	snapshotservice "prostic/internal/service/snapshots"
)

type RefreshResult struct {
	SnapshotCount int `json:"snapshotCount"`
}

func RefreshAll() (*RefreshResult, error) {
	snapshotCount, err := snapshotservice.RefreshSnapshotCache()
	if err != nil {
		return nil, err
	}

	if err := repostatsservice.RefreshRepoStatCache(); err != nil {
		return nil, err
	}

	return &RefreshResult{
		SnapshotCount: snapshotCount,
	}, nil
}
