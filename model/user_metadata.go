package model

import (
	"time"
)

type UserMetadata struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;unique" json:"user_id"`
	Phone     string    `gorm:"size:20" json:"phone"`
	State     int       `gorm:"not null;default:1" json:"state"` // 1=activo, 0=inactivo, etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserMetadata) TableName() string {
	return "user_metadata"
}
