package dto

import "github.com/go-playground/validator/v10"

// ProfileSummary para respuestas anidadas
type ProfileSummary struct {
  ID   uint   `json:"id"`
  Name string `json:"name"`
}

// UserCreateRequest para POST /users
type UserCreateRequest struct {
  Name      string `json:"name" binding:"required,min=2,max=255" example:"César Pérez"`
  Email     string `json:"email" binding:"required,email,max=255" example:"cesar@tudominio.com"`
  Password  string `json:"password" binding:"required,min=8,max=100" example:"********"`
  // Metadata (opcional)
  Phone     string `json:"phone" binding:"omitempty,max=20" example:"+56912345678"`
  State     *int   `json:"state" binding:"omitempty,oneof=0 1" example:"1"`
  ProfileID uint   `json:"profile_id" binding:"omitempty" example:"1"`
}

func (u UserCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  errs := make(map[string]string)
  for _, e := range ve {
    switch e.Field() {
    case "Name":
      switch e.Tag() {
      case "required":
        errs["name"] = "El nombre es obligatorio"
      case "min":
        errs["name"] = "El nombre debe tener al menos 2 caracteres"
      case "max":
        errs["name"] = "El nombre no puede exceder 255 caracteres"
      }
    case "Email":
      switch e.Tag() {
      case "required":
        errs["email"] = "El email es obligatorio"
      case "email":
        errs["email"] = "El email no tiene un formato válido"
      case "max":
        errs["email"] = "El email no puede exceder 255 caracteres"
      }
    case "Password":
      switch e.Tag() {
      case "required":
        errs["password"] = "La contraseña es obligatoria"
      case "min":
        errs["password"] = "La contraseña debe tener al menos 8 caracteres"
      case "max":
        errs["password"] = "La contraseña no puede exceder 100 caracteres"
      }
    case "Phone":
      if e.Tag() == "max" {
        errs["phone"] = "El teléfono no puede exceder 20 caracteres"
      }
    case "State":
      if e.Tag() == "oneof" {
        errs["state"] = "El estado debe ser 0 (inactivo) o 1 (activo)"
      }
    }
  }
  return errs
}

// UserUpdateRequest para PUT /users/:id
type UserUpdateRequest struct {
  Name      string `json:"name" binding:"omitempty,min=2,max=255" example:"César Pérez"`
  Email     string `json:"email" binding:"omitempty,email,max=255" example:"cesar@tudominio.com"`
  Password  string `json:"password" binding:"omitempty,min=8,max=100" example:"********"`
  Phone     string `json:"phone" binding:"omitempty,max=20" example:"+56912345678"`
  State     *int   `json:"state" binding:"omitempty,oneof=0 1" example:"1"`
  ProfileID uint   `json:"profile_id" binding:"omitempty" example:"1"`
}

func (u UserUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  return UserCreateRequest(u).MensajesDeError(ve)
}

// UserResponse para respuestas (⚠️ SIN PASSWORD, con date/time formateados)
type UserResponse struct {
  ID        uint                   `json:"id"`
  Name      string                 `json:"name"`
  Email     string                 `json:"email"`
  Date      string                 `json:"date"`
  Time      string                 `json:"time"`
  Phone     string                 `json:"phone,omitempty"`
  State     int                    `json:"state"`
  ProfileID uint                   `json:"profile_id,omitempty"`
  Profile   *ProfileSummary        `json:"profile,omitempty"`
  Modules   []UserModuleResponse   `json:"modules,omitempty"` // ← Usa el tipo específico
}

type UsersResponse []UserResponse

// ✅ UserModuleResponse: representa un módulo con sus items/permisos para el usuario
// (Diferente de ModuleResponse que es para CRUD de módulos)
type UserModuleResponse struct {
  Name  string         `json:"name"`
  Slug  string         `json:"slug"`
  Items []ItemResponse `json:"items"` // ← Items con code incluido
}