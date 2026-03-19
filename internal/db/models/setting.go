package models

import "time"

type Setting struct {
	ID                  uint   `gorm:"primaryKey"`
	PasswordHash        string `gorm:"not null"`
	NeedsPasswordChange bool   `gorm:"not null;default:false"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
