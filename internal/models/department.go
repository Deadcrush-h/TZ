package models

import (
	"time"
)

type Department struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:200;not null" json:"name"`
	ParentID  *uint     `gorm:"index" json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (Department) TableName() string {
	return "departments"
}
