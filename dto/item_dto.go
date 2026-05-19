package dto

import "github.com/go-playground/validator/v10"

// ItemCreateRequest para POST /items
type ItemCreateRequest struct {
	// ✅ Validación simple: required + longitud
	// Si necesitas formato slug estricto, valida manualmente en el handler
	Name string `json:"name" binding:"required,min=3,max=50" example:"crear_usuario"`
}

func (i ItemCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
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
				errs["name"] = "El nombre no puede exceder 50 caracteres"
			}
		}
	}
	return errs
}

// ItemUpdateRequest para PUT /items/:id
type ItemUpdateRequest struct {
	Name string `json:"name" binding:"required,min=3,max=50" example:"actualizar_usuario"`
}

func (i ItemUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	return ItemCreateRequest{Name: i.Name}.MensajesDeError(ve)
}

// ItemResponse para respuestas
type ItemResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// ItemsResponse para listas
type ItemsResponse []ItemResponse