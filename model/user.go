package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uint           `gorm:"primaryKey;autoIncrement;type:bigint" json:"id"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	Email          string         `gorm:"size:255;not null;unique" json:"email"`
	Password       string         `gorm:"size:255;not null" json:"-"` // Ocultar en JSON
	RememberToken  string         `gorm:"size:100" json:"-"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}

// HashPassword genera el hash de la contraseña
func (u *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

// CheckPassword verifica si la contraseña coincide
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
