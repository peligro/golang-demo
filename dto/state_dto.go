package dto

import (
	"github.com/go-playground/validator/v10"
)

// StateCreateRequest para POST /states
type StateCreateRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100" example:"Activo"`
}

func (s StateCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	errs := make(map[string]string)
	for _, e := range ve {
		switch e.Field() {
		case "Name":
			switch e.Tag() {
			case "required":
				errs["name"] = "El nombre es obligatorio"
			case "min":
				errs["name"] = "El nombre debe tener al menos 3 caracteres"
			case "max":
				errs["name"] = "El nombre no puede exceder 100 caracteres"
			}
		}
	}
	return errs
}

// StateUpdateRequest para PUT /states/:id
type StateUpdateRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100" example:"Activo"`
}

func (s StateUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	return StateCreateRequest{Name: s.Name}.MensajesDeError(ve)
}

// StateResponse para respuestas
type StateResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// StatesResponse para listas
type StatesResponse []StateResponse
