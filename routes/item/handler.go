package item

import (
  "errors" 
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// Handler maneja las operaciones HTTP para Items
type Handler struct {
  service *Service
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
  return &Handler{
    service: NewService(db), // ← Inyectar servicio
  }
}

// Index godoc
// @Summary Listar items
// @Description Retorna todos los items registrados ordenados por ID descendente
// @Tags Items
// @Produce json
// @Success 200 {array} dto.ItemResponse
// @Router /items [get]
func (h *Handler) Index(c *gin.Context) {
  items, err := h.service.ListAll()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar items",
    })
    return
  }
  c.JSON(http.StatusOK, items)
}

// Show godoc
// @Summary Obtener item por ID
// @Description Retorna un item específico por su ID
// @Tags Items
// @Produce json
// @Param id path uint true "ID del item"
// @Success 200 {object} dto.ItemResponse
// @Failure 404 {object} map[string]string
// @Router /items/{id} [get]
func (h *Handler) Show(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  item, err := h.service.GetByID(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar el item",
    })
    return
  }
  c.JSON(http.StatusOK, item)
}

// Create godoc
// @Summary Crear nuevo item
// @Description Crea un nuevo item (acción) con nombre y código únicos
// @Tags Items
// @Accept json
// @Produce json
// @Param item body dto.ItemCreateRequest true "Datos del item"
// @Success 201 {object} dto.ItemResponse
// @Failure 400 {object} map[string]string
// @Router /items [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.ItemCreateRequest](c)
  if !ok {
    return
  }

  item, err := h.service.Create(body.Name, body.Code)
  if err != nil {
    if err.Error() == "nombre duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": common.ErrDuplicate.Error(),
        "field":   "name",
      })
      return
    }
    if err.Error() == "código duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "El código ya está en uso",
        "field":   "code",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al crear el item",
    })
    return
  }
  c.JSON(http.StatusCreated, item)
}

// Update godoc
// @Summary Actualizar item
// @Description Actualiza el nombre y código de un item existente
// @Tags Items
// @Accept json
// @Produce json
// @Param id path uint true "ID del item"
// @Param item body dto.ItemUpdateRequest true "Nuevos datos"
// @Success 200 {object} dto.ItemResponse
// @Failure 404 {object} map[string]string
// @Router /items/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  body, ok := common.BindAndValidate[dto.ItemUpdateRequest](c)
  if !ok {
    return
  }

  item, err := h.service.Update(uint(id), body.Name, body.Code)
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    if err.Error() == "nombre duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": common.ErrDuplicate.Error(),
        "field":   "name",
      })
      return
    }
    if err.Error() == "código duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "El código ya está en uso",
        "field":   "code",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al actualizar el item",
    })
    return
  }
  c.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary Eliminar item
// @Description Elimina un item por su ID (valida dependencias con profile_module_item)
// @Tags Items
// @Produce json
// @Param id path uint true "ID del item"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /items/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  err := h.service.Delete(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    if errors.Is(err, common.ErrHasDependencies) {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": common.ErrHasDependencies.Error(),
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al eliminar el item",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Item eliminado exitosamente",
  })
}