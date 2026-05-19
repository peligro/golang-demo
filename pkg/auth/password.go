package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword genera un hash bcrypt para la contraseña
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compara una contraseña en texto plano con un hash bcrypt
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
