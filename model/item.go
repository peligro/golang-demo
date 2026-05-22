package model

import (
	"time"
)

type Item struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Code      string    `gorm:"size:50;not null;uniqueIndex" json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Item) TableName() string {
	return "item"
}
