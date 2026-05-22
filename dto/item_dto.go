package dto

import "github.com/go-playground/validator/v10"

type ItemCreateRequest struct {
  Name string `json:"name" binding:"required,min=3,max=50" example:"Crear Usuario"`
  Code string `json:"code" binding:"required,slug,min=3,max=50" example:"crear_usuario"`
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
    case "Code":
      switch e.Tag() {
      case "required":
        errs["code"] = "El código es obligatorio"
      case "slug":
        errs["code"] = "El código solo puede contener letras minúsculas, números y guiones bajos"
      case "min":
        errs["code"] = "El código debe tener al menos 3 caracteres"
      case "max":
        errs["code"] = "El código no puede exceder 50 caracteres"
      }
    }
  }
  return errs
}

type ItemUpdateRequest struct {
  Name string `json:"name" binding:"required,min=3,max=50" example:"Actualizar Usuario"`
  Code string `json:"code" binding:"required,slug,min=3,max=50" example:"actualizar_usuario"`
}

func (i ItemUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  return ItemCreateRequest{Name: i.Name, Code: i.Code}.MensajesDeError(ve)
}

type ItemResponse struct {
  ID   uint   `json:"id"`
  Name string `json:"name"`
  Code string `json:"code"`
}

type ItemsResponse []ItemResponse