package model

import "time"

type AppMenu struct {
  ID        uint      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
  Label     string    `gorm:"size:200;column:label" json:"label"`
  Title     string    `gorm:"size:200;column:title" json:"title"`
  Icon      string    `gorm:"size:200;column:icon" json:"icon"`
  Order     int       `gorm:"column:order" json:"order"`
  ParentID  *uint     `gorm:"column:parentId" json:"parent_id"`      // ← Pointer (opcional)
  ModuleID  *uint     `gorm:"column:moduleId;index" json:"module_id"` // ← Pointer (opcional) + índice
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

func (AppMenu) TableName() string {
  return "app_menu"
}