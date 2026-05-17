package model

import (
	"time"
)

type ProfileModuleItem struct {
	ID               uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	ProfileModuleID  uint            `gorm:"not null;index:idx_pm_item,unique" json:"profile_module_id"`
	ItemID           uint            `gorm:"not null;index:idx_pm_item,unique" json:"item_id"`
	ProfileModule    *ProfileModule  `gorm:"foreignKey:ProfileModuleID" json:"profile_module,omitempty"`
	Item             *Item           `gorm:"foreignKey:ItemID" json:"item,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (ProfileModuleItem) TableName() string {
	return "profile_module_item"
}
