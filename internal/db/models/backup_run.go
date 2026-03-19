package models

import "time"

type BackupRun struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	BackupID       string     `gorm:"index" json:"backupID"`
	Trigger        string     `gorm:"index;not null" json:"trigger"`
	Status         string     `gorm:"index;not null" json:"status"`
	Logs           string     `gorm:"type:text" json:"logs"`
	TotalItems     int        `gorm:"not null;default:0" json:"totalItems"`
	CompletedItems int        `gorm:"not null;default:0" json:"completedItems"`
	StartedAt      time.Time  `gorm:"index;not null" json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}
