package dto

import "github.com/go-playground/validator/v10"

// LoginRequest para POST /auth/login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"cesar@tudominio.com"`
	Password string `json:"password" binding:"required,min=8" example:"********"` // ← Placeholder genérico

}

func (l LoginRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)
	for _, e := range ve {
		switch e.Field() {
		case "Email":
			switch e.Tag() {
			case "required":
				errs["email"] = "El email es obligatorio"
			case "email":
				errs["email"] = "El email no tiene un formato válido"
			}
		case "Password":
			switch e.Tag() {
			case "required":
				errs["password"] = "La contraseña es obligatoria"
			case "min":
				errs["password"] = "La contraseña debe tener al menos 8 caracteres"
			}
		}
	}
	return errs
}

// LoginResponse respuesta del login (SIN tokens en el body)
type LoginResponse struct {
	User UserResponse `json:"user"`
}

// MeResponse respuesta de /auth/me
type MeResponse struct {
	User UserResponse `json:"user"`
}

// RefreshResponse respuesta de /auth/refresh
type RefreshResponse struct {
	Message string `json:"message"`
}

// LogoutResponse respuesta de /auth/logout
type LogoutResponse struct {
	Message string `json:"message"`
}
