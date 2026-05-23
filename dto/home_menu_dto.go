package dto

import (
  "time"
  "github.com/go-playground/validator/v10"
)

type HomeMenuCreateRequest struct {
  Title       string `json:"title" binding:"required,max=200" example:"Dashboard"`
  Icon        string `json:"icon" binding:"required,max=100" example:"fa-chart-line"`
  Color       string `json:"color" binding:"required,max=100" example:"#3b82f6"`
  Description string `json:"description" binding:"required" example:"Panel principal de control"`
  Slug        string `json:"slug" binding:"required,max=200" example:"/dashboard"`
  Order       int    `json:"order" binding:"omitempty" example:"1"`
  ModuleID    *uint  `json:"module_id" binding:"omitempty" example:"1"`
}

func (h HomeMenuCreateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  errs := make(map[string]string)
  for _, e := range ve {
    switch e.Field() {
    case "Title":
      errs["title"] = "El título es obligatorio y no puede exceder 200 caracteres"
    case "Icon":
      errs["icon"] = "El ícono es obligatorio y no puede exceder 100 caracteres"
    case "Color":
      errs["color"] = "El color es obligatorio y no puede exceder 100 caracteres"
    case "Description":
      errs["description"] = "La descripción es obligatoria"
    case "Slug":
      errs["slug"] = "El slug es obligatorio y no puede exceder 200 caracteres"
    }
  }
  return errs
}

type HomeMenuUpdateRequest struct {
  Title       string `json:"title" binding:"omitempty,max=200" example:"Dashboard"`
  Icon        string `json:"icon" binding:"omitempty,max=100" example:"fa-chart-line"`
  Color       string `json:"color" binding:"omitempty,max=100" example:"#3b82f6"`
  Description string `json:"description" binding:"omitempty" example:"Panel principal de control"`
  Slug        string `json:"slug" binding:"omitempty,max=200" example:"/dashboard"`
  Order       int    `json:"order" binding:"omitempty" example:"1"`
  ModuleID    *uint  `json:"module_id" binding:"omitempty" example:"1"`
}

func (h HomeMenuUpdateRequest) MensajesDeError(ve validator.ValidationErrors) map[string]string {
  return HomeMenuCreateRequest{
    Title: h.Title, Icon: h.Icon, Color: h.Color,
    Description: h.Description, Slug: h.Slug,
    Order: h.Order, ModuleID: h.ModuleID,
  }.MensajesDeError(ve)
}

type HomeMenuResponse struct {
  ID          uint      `json:"id"`
  Title       string    `json:"title"`
  Icon        string    `json:"icon"`
  Color       string    `json:"color"`
  Description string    `json:"description"`
  Slug        string    `json:"slug"`
  Order       int       `json:"order"`
  ModuleID    *uint     `json:"module_id"`
  CreatedAt   time.Time `json:"created_at"`
  UpdatedAt   time.Time `json:"updated_at"`
}

type HomeMenusResponse []HomeMenuResponse
