package model

import (
  "time"
)

type Module struct {
  ID                 uint                `gorm:"primaryKey;autoIncrement" json:"id"`
  Name               string              `gorm:"size:100;not null;unique" json:"name"`
  Slug               string              `gorm:"size:100;not null;unique" json:"slug"`
  CreatedAt          time.Time           `json:"created_at"`
  UpdatedAt          time.Time           `json:"updated_at"`
}

func (Module) TableName() string {
  return "module"
}