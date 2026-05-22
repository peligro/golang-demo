package profile

import (
  "errors"
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// Handler maneja las operaciones HTTP para Profiles
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
// @Summary Listar perfiles con paginación y búsqueda
// @Description Retorna perfiles paginados, con búsqueda por nombre y ordenamiento
// @Tags Profiles
// @Produce json
// @Param page query int false "Página" default(1)
// @Param per_page query int false "Registros por página" default(20)
// @Param search query string false "Término de búsqueda"
// @Param field query string false "Campo a buscar" default(name) Enums(name)
// @Param sort_by query string false "Campo para ordenar" Enums(id,name,created_at)
// @Param sort_dir query string false "Dirección" Enums(asc,desc)
// @Success 200 {object} common.PaginatedResponse[dto.ProfileResponse]
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles [get]
func (h *Handler) Index(c *gin.Context) {
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
      "message": "Error al consultar perfiles",
    })
    return
  }
  c.JSON(http.StatusOK, response)
}

// Show godoc
// @Summary Obtener perfil por ID
// @Description Retorna un perfil específico por su ID
// @Tags Profiles
// @Produce json
// @Param id path uint true "ID del perfil"
// @Success 200 {object} dto.ProfileResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id} [get]
func (h *Handler) Show(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  profile, err := h.service.GetByID(uint(id))
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
      "message": "Error al consultar el perfil",
    })
    return
  }
  c.JSON(http.StatusOK, profile)
}

// Create godoc
// @Summary Crear nuevo perfil
// @Description Crea un nuevo perfil con nombre y descripción
// @Tags Profiles
// @Accept json
// @Produce json
// @Param profile body dto.ProfileCreateRequest true "Datos del perfil"
// @Success 201 {object} dto.ProfileResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.ProfileCreateRequest](c)
  if !ok {
    return
  }

  profile, err := h.service.Create(body.Name, body.Description)
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
      "message": "Error al crear el perfil",
    })
    return
  }
  c.JSON(http.StatusCreated, profile)
}

// Update godoc
// @Summary Actualizar perfil
// @Description Actualiza nombre y descripción de un perfil existente
// @Tags Profiles
// @Accept json
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param profile body dto.ProfileUpdateRequest true "Nuevos datos"
// @Success 200 {object} dto.ProfileResponse
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok {
    return
  }

  body, ok := common.BindAndValidate[dto.ProfileUpdateRequest](c)
  if !ok {
    return
  }

  profile, err := h.service.Update(uint(id), body.Name, body.Description)
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
      "message": "Error al actualizar el perfil",
    })
    return
  }
  c.JSON(http.StatusOK, profile)
}

// Delete godoc
// @Summary Eliminar perfil
// @Description Elimina un perfil por su ID (valida dependencias con profile_module)
// @Tags Profiles
// @Produce json
// @Param id path uint true "ID del perfil"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id} [delete]
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
      "message": "Error al eliminar el perfil",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Perfil eliminado exitosamente",
  })
}