package models

import "time"

type Task struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Purpose    string     `gorm:"index;not null" json:"purpose"`
	Status     string     `gorm:"index;not null" json:"status"`
	Logs       string     `gorm:"type:text" json:"logs"`
	StartedAt  time.Time  `gorm:"index;not null" json:"startedAt"`
	FinishedAt *time.Time `json:"finishedAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
