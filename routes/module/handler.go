package module

import (
  "errors"
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// Handler maneja las operaciones HTTP para Modules
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
// @Summary Listar módulos con paginación y búsqueda
// @Description Retorna módulos paginados, con búsqueda por nombre y ordenamiento
// @Tags Modules
// @Produce json
// @Param page query int false "Página" default(1)
// @Param per_page query int false "Registros por página" default(20)
// @Param search query string false "Término de búsqueda"
// @Param field query string false "Campo a buscar" default(name) Enums(name)
// @Param sort_by query string false "Campo para ordenar" Enums(id,name,slug,created_at)
// @Param sort_dir query string false "Dirección" Enums(asc,desc)
// @Success 200 {object} common.PaginatedResponse[dto.ModuleResponse]
// @Router /modules [get]
func (h *Handler) Index(c *gin.Context) {
  // Parsear parámetros con valores por defecto
  params := common.DefaultPagination()
  if err := c.ShouldBindQuery(&params); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "Parámetros de consulta inválidos",
    })
    return
  }

  response, err := h.service.ListPaginated(params)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar módulos",
    })
    return
  }
  c.JSON(http.StatusOK, response)
}

// Show godoc
// @Summary Obtener módulo por ID
// @Description Retorna un módulo específico por su ID
// @Tags Modules
// @Produce json
// @Param id path uint true "ID del módulo"
// @Success 200 {object} dto.ModuleResponse
// @Failure 404 {object} map[string]string
// @Router /modules/{id} [get]
func (h *Handler) Show(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  module, err := h.service.GetByID(uint(id))
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
      "message": "Error al consultar el módulo",
    })
    return
  }
  c.JSON(http.StatusOK, module)
}

// Create godoc
// @Summary Crear nuevo módulo
// @Description Crea un nuevo módulo con nombre y path (slug)
// @Tags Modules
// @Accept json
// @Produce json
// @Param module body dto.ModuleCreateRequest true "Datos del módulo"
// @Success 201 {object} dto.ModuleResponse
// @Failure 400 {object} map[string]string
// @Router /modules [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.ModuleCreateRequest](c)
  if !ok {
    return
  }

  module, err := h.service.Create(body.Name, body.Slug)
  if err != nil {
    if err.Error() == "nombre duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": common.ErrDuplicate.Error(),
        "field":   "name",
      })
      return
    }
    if err.Error() == "path duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "El path ya está en uso",
        "field":   "slug",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al crear el módulo",
    })
    return
  }
  c.JSON(http.StatusCreated, module)
}

// Update godoc
// @Summary Actualizar módulo
// @Description Actualiza nombre y path de un módulo existente
// @Tags Modules
// @Accept json
// @Produce json
// @Param id path uint true "ID del módulo"
// @Param module body dto.ModuleUpdateRequest true "Nuevos datos"
// @Success 200 {object} dto.ModuleResponse
// @Failure 404 {object} map[string]string
// @Router /modules/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  body, ok := common.BindAndValidate[dto.ModuleUpdateRequest](c)
  if !ok {
    return
  }

  module, err := h.service.Update(uint(id), body.Name, body.Slug)
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
    if err.Error() == "path duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "El path ya está en uso",
        "field":   "slug",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al actualizar el módulo",
    })
    return
  }
  c.JSON(http.StatusOK, module)
}

// Delete godoc
// @Summary Eliminar módulo
// @Description Elimina un módulo por su ID (valida dependencias con profile_module)
// @Tags Modules
// @Produce json
// @Param id path uint true "ID del módulo"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /modules/{id} [delete]
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
      "message": "Error al eliminar el módulo",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Módulo eliminado exitosamente",
  })
}