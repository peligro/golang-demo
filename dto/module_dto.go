package dto

import "github.com/go-playground/validator/v10"

// ModuleCreateRequest para POST /modules
type ModuleCreateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100" example:"Usuarios"`
	Description string `json:"description" binding:"required,min=10,max=500" example:"Gestión de usuarios del sistema"`
}

func (m ModuleCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
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
		case "Description":
			switch e.Tag() {
			case "required":
				errs["description"] = "La descripción es obligatoria"
			case "min":
				errs["description"] = "La descripción debe tener al menos 10 caracteres"
			case "max":
				errs["description"] = "La descripción no puede exceder 500 caracteres"
			}
		}
	}
	return errs
}

// ModuleUpdateRequest para PUT /modules/:id
type ModuleUpdateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100" example:"Usuarios"`
	Description string `json:"description" binding:"required,min=10,max=500" example:"Gestión de usuarios del sistema"`
}

func (m ModuleUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	return ModuleCreateRequest{Name: m.Name, Description: m.Description}.MensajesDeError(ve)
}

// ModuleResponse para respuestas
type ModuleResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ModulesResponse para listas
type ModulesResponse []ModuleResponse