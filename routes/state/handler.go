package state

import (
  "errors"
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// Handler maneja las operaciones HTTP para States
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
// @Summary Listar estados
// @Description Retorna todos los estados registrados ordenados por ID descendente
// @Tags States
// @Produce json
// @Success 200 {array} dto.StateResponse
// @Failure 500 {object} map[string]string
// @Router /states [get]
func (h *Handler) Index(c *gin.Context) {
  states, err := h.service.ListAll()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar estados",
    })
    return
  }
  c.JSON(http.StatusOK, states)
}

// Show godoc
// @Summary Obtener estado por ID
// @Description Retorna un estado específico por su ID
// @Tags States
// @Produce json
// @Param id path uint true "ID del estado" minimum(1)
// @Success 200 {object} dto.StateResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /states/{id} [get]
func (h *Handler) Show(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  state, err := h.service.GetByID(uint(id))
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
      "message": "Error al consultar el estado",
    })
    return
  }
  c.JSON(http.StatusOK, state)
}

// Create godoc
// @Summary Crear nuevo estado
// @Description Crea un nuevo estado con nombre único
// @Tags States
// @Accept json
// @Produce json
// @Param state body dto.StateCreateRequest true "Datos del estado"
// @Success 201 {object} dto.StateResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /states [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.StateCreateRequest](c)
  if !ok {
    return
  }

  state, err := h.service.Create(body.Name)
  if err != nil {
    if err.Error() == "nombre duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": common.ErrDuplicate.Error(),
        "field":   "name",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al crear el estado",
    })
    return
  }
  c.JSON(http.StatusCreated, state)
}

// Update godoc
// @Summary Actualizar estado
// @Description Actualiza el nombre de un estado existente
// @Tags States
// @Accept json
// @Produce json
// @Param id path uint true "ID del estado" minimum(1)
// @Param state body dto.StateUpdateRequest true "Nuevo nombre"
// @Success 200 {object} dto.StateResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /states/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  body, ok := common.BindAndValidate[dto.StateUpdateRequest](c)
  if !ok {
    return
  }

  state, err := h.service.Update(uint(id), body.Name)
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
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al actualizar el estado",
    })
    return
  }
  c.JSON(http.StatusOK, state)
}

// Delete godoc
// @Summary Eliminar estado
// @Description Elimina un estado por su ID (valida dependencias con user_metadata)
// @Tags States
// @Produce json
// @Param id path uint true "ID del estado" minimum(1)
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /states/{id} [delete]
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
      "message": "Error al eliminar el estado",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Estado eliminado exitosamente",
  })
}