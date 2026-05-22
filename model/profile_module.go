package model

import "time"

type ProfileModule struct {
  ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
  
  ModuleID    uint      `gorm:"not null;index:idx_profile_module,unique" json:"module_id"`
  Module      *Module   `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
  ProfileID   uint      `gorm:"not null;index:idx_profile_module,unique" json:"profile_id"`
  Profile     *Profile  `gorm:"foreignKey:ProfileID" json:"profile,omitempty"`
  CreatedAt   time.Time `json:"created_at"`
  UpdatedAt   time.Time `json:"updated_at"`
}

func (ProfileModule) TableName() string {
  return "profile_module"
}