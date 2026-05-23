package dto

import (
	"time"
	"github.com/go-playground/validator/v10"
)

type AppMenuCreateRequest struct {
  Label    string `json:"label" binding:"required,max=200" example:"Inicio"`
  Title    string `json:"title" binding:"required,max=200" example:"Página principal"`
  Icon     string `json:"icon" binding:"omitempty,max=200" example:"fa-home"`
  Order    int    `json:"order" binding:"omitempty" example:"1"`
  ParentID *uint  `json:"parent_id" binding:"omitempty" example:"0"`
  ModuleID *uint  `json:"module_id" binding:"omitempty" example:"1"`
}

func (a AppMenuCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  errs := make(map[string]string)
  for _, e := range ve {
    switch e.Field() {
    case "Label":
      errs["label"] = "La etiqueta es obligatoria y no puede exceder 200 caracteres"
    case "Title":
      errs["title"] = "El título es obligatorio y no puede exceder 200 caracteres"
    case "Icon":
      errs["icon"] = "El ícono no puede exceder 200 caracteres"
    }
  }
  return errs
}

type AppMenuUpdateRequest struct {
  Label    string `json:"label" binding:"omitempty,max=200" example:"Inicio"`
  Title    string `json:"title" binding:"omitempty,max=200" example:"Página principal"`
  Icon     string `json:"icon" binding:"omitempty,max=200" example:"fa-home"`
  Order    int    `json:"order" binding:"omitempty" example:"1"`
  ParentID *uint  `json:"parent_id" binding:"omitempty" example:"0"`
  ModuleID *uint  `json:"module_id" binding:"omitempty" example:"1"`
}

func (a AppMenuUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  return AppMenuCreateRequest{
    Label: a.Label, Title: a.Title, Icon: a.Icon,
    Order: a.Order, ParentID: a.ParentID, ModuleID: a.ModuleID,
  }.MensajesDeError(ve)
}

type AppMenuResponse struct {
  ID        uint      `json:"id"`
  Label     string    `json:"label"`
  Title     string    `json:"title"`
  Icon      string    `json:"icon"`
  Order     int       `json:"order"`
  ParentID  *uint     `json:"parent_id"`
  ModuleID  *uint     `json:"module_id"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

type AppMenusResponse []AppMenuResponse
