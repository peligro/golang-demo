package model

import (
	"time"
)

type UserMetadata struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;unique;index" json:"user_id"`
	Phone     string    `gorm:"size:20" json:"phone"`
	State     int       `gorm:"not null;default:1;index" json:"state"` // 1=activo, 0=inactivo
	ProfileID uint      `gorm:"default:0;index" json:"profile_id"`     // ← FK con índice para performance
	
	// ← Relación GORM para Preload
	Profile   *Profile  `gorm:"foreignKey:ProfileID;references:ID" json:"profile,omitempty"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserMetadata) TableName() string {
	return "user_metadata"
}