package model

import (
	"time"
)

type Profile struct {
	ID          uint             `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string           `gorm:"size:100;not null;unique" json:"name"`
	Description string           `gorm:"type:text" json:"description"`
	Modules     []ProfileModule  `gorm:"foreignKey:ProfileID" json:"modules,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

func (Profile) TableName() string {
	return "profile"
}
