package model

import (
    "fmt"
    "gorm.io/gorm"
)

func Migrations(db *gorm.DB) error {
    err := db.AutoMigrate(
        &State{}, &User{}, &UserMetadata{},
        &Module{}, &Item{}, &Profile{},
        &ProfileModule{}, &ProfileModuleItem{}, &AppMenu{}, &HomeMenu{},
    )
    if err != nil {
        return fmt.Errorf("error en migraciones: %w", err)
    }
    return nil
}