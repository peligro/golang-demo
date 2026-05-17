package model

import (
	"time"
)

type ProfileModule struct {
	ID             uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	ProfileID      uint                `gorm:"not null;index:idx_profile_module,unique" json:"profile_id"`
	ModuleID       uint                `gorm:"not null;index:idx_profile_module,unique" json:"module_id"`
	Profile        *Profile            `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
	Module         *Module             `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Items          []ProfileModuleItem `gorm:"foreignKey:ProfileModuleID" json:"items,omitempty"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

func (ProfileModule) TableName() string {
	return "profile_module"
}
