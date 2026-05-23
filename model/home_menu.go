package model

import "time"

type HomeMenu struct {
  ID          uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
  Title       string    `gorm:"size:200;not null;column:title" json:"title"`
  Icon        string    `gorm:"size:100;not null;column:icon" json:"icon"`
  Color       string    `gorm:"size:100;not null;column:color" json:"color"`
  Description string    `gorm:"type:text;not null;column:description" json:"description"`
  Slug        string    `gorm:"size:200;not null;default:'vacío';column:slug" json:"slug"`
  Order       int       `gorm:"not null;default:1;column:order" json:"order"`
  ModuleID    *uint     `gorm:"column:moduleId;index" json:"module_id"`
  CreatedAt   time.Time `json:"created_at"`
  UpdatedAt   time.Time `json:"updated_at"`
}

func (HomeMenu) TableName() string {
  return "home_menu"
}
