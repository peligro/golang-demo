package dto

import "github.com/go-playground/validator/v10"

// ProfileCreateRequest para POST /profiles
type ProfileCreateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100" example:"Administrador"`
	Description string `json:"description" binding:"required,min=10,max=500" example:"Acceso completo a todos los módulos del sistema"`
}

func (p ProfileCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
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

// ProfileUpdateRequest para PUT /profiles/:id
type ProfileUpdateRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100" example:"Administrador"`
	Description string `json:"description" binding:"required,min=10,max=500" example:"Acceso completo a todos los módulos del sistema"`
}

func (p ProfileUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
	return ProfileCreateRequest{Name: p.Name, Description: p.Description}.MensajesDeError(ve)
}

// ProfileResponse para respuestas
type ProfileResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProfilesResponse para listas
type ProfilesResponse []ProfileResponse


// =============================================================================
// PROFILE MODULE REQUESTS
// =============================================================================

// ProfileModuleSyncRequest para PUT /profiles/:id/modules
type ProfileModuleSyncRequest struct {
	Modules []uint `json:"modules" binding:"omitempty"`
}

// ProfileModuleItemSyncRequest para PUT /profiles/:id/modules/:moduleId/items
type ProfileModuleItemSyncRequest struct {
	Items []uint `json:"items" binding:"omitempty"`
}

// ProfileModuleItemAttachRequest para POST /profiles/:id/modules/:moduleId/items
type ProfileModuleItemAttachRequest struct {
	ItemID uint `json:"item_id" binding:"required" example:"1"`
}