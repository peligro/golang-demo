package profile

import (
  "errors"
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// ModuleHandler maneja las operaciones HTTP para ProfileModule
type ModuleHandler struct {
  service *ModuleService
}

// NewModuleHandler crea una nueva instancia
func NewModuleHandler(db *gorm.DB) *ModuleHandler {
  return &ModuleHandler{
    service: NewModuleService(db), // ← Inyectar servicio
  }
}

// Index godoc
// @Summary Listar módulos asignados a un perfil
// @Description Retorna los módulos asignados a un perfil específico
// @Tags Profile Modules
// @Produce json
// @Param id path uint true "ID del perfil"
// @Success 200 {array} dto.ModuleResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/modules [get]
func (h *ModuleHandler) Index(c *gin.Context) {
  profileID, ok := common.ParseID(c, "id")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de perfil inválido",
    })
    return
  }

  modules, err := h.service.ListByProfile(uint(profileID))
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
      "message": "Error al consultar módulos",
    })
    return
  }
  c.JSON(http.StatusOK, modules)
}

// Sync godoc
// @Summary Sincronizar módulos de un perfil
// @Description Asigna o remueve módulos de un perfil (operación sync)
// @Tags Profile Modules
// @Accept json
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param modules body dto.ProfileModuleSyncRequest true "Array de IDs de módulos"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/modules [put]
func (h *ModuleHandler) Sync(c *gin.Context) {
  profileID, ok := common.ParseID(c, "id")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de perfil inválido",
    })
    return
  }

  var req dto.ProfileModuleSyncRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "Datos inválidos",
    })
    return
  }

  err := h.service.Sync(uint(profileID), req.Modules)
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    if err.Error() == "uno o más módulos no existen" {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "Uno o más módulos no existen",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al sincronizar módulos",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Módulos actualizados exitosamente",
  })
}