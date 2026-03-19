package models

import "time"

type Snapshot struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SnapshotID   string    `gorm:"uniqueIndex;not null" json:"snapshotID"`
	Time         time.Time `gorm:"index" json:"time"`
	Hostname     string    `json:"hostname"`
	Tree         string    `json:"tree"`
	Paths        string    `gorm:"type:text" json:"paths"`
	Tags         string    `gorm:"type:text" json:"tags"`
	BackupID     string    `gorm:"index" json:"backupID"`
	VMID         *int      `gorm:"index" json:"vmid"`
	Name         string    `json:"name"`
	VMType       string    `json:"vmType"`
	SnapshotType string    `gorm:"index" json:"snapshotType"`
	BackupDate   string    `json:"backupDate"`
	DestFile     string    `gorm:"type:text" json:"destFile"`
	SrcFile      string    `gorm:"type:text" json:"srcFile"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
