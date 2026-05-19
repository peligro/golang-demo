package dto

import "github.com/go-playground/validator/v10"

// ModuleCreateRequest para POST /modules
type ModuleCreateRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100" example:"Usuarios"`
	Slug string `json:"slug" binding:"required,startswith=/,max=100" example:"/setting/users"`
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
		case "Slug":
			switch e.Tag() {
			case "required":
				errs["slug"] = "El path es obligatorio"
			case "startswith":
				errs["slug"] = "El path debe comenzar con /"
			case "max":
				errs["slug"] = "El path no puede exceder 100 caracteres"
			}
		}
	}
	return errs
}

// ModuleUpdateRequest para PUT /modules/:id
type ModuleUpdateRequest struct {
	Name string `json:"name" binding:"required,min=3,max=100" example:"Usuarios"`
	Slug string `json:"slug" binding:"required,startswith=/,max=100" example:"/setting/users"`
}

func (m ModuleUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	return ModuleCreateRequest{Name: m.Name, Slug: m.Slug}.MensajesDeError(ve)
}

// ModuleResponse para respuestas (compatible con frontend)
type ModuleResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}