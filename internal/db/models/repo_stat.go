package models

import "time"

type RepoStat struct {
	ID                     uint      `gorm:"primaryKey" json:"id"`
	TotalSize              int64     `json:"totalSize"`
	TotalUncompressedSize  int64     `json:"totalUncompressedSize"`
	CompressionRatio       float64   `json:"compressionRatio"`
	CompressionProgress    int       `json:"compressionProgress"`
	CompressionSpaceSaving float64   `json:"compressionSpaceSaving"`
	TotalBlobCount         int64     `json:"totalBlobCount"`
	SnapshotsCount         int64     `json:"snapshotsCount"`
	LastRefreshedAt        time.Time `json:"lastRefreshedAt"`
	CreatedAt              time.Time `json:"createdAt"`
	UpdatedAt              time.Time `json:"updatedAt"`
}
