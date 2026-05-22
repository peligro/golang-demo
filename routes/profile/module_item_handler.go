package profile

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// ModuleItemHandler maneja las operaciones HTTP para ProfileModuleItem
type ModuleItemHandler struct {
  service *ModuleItemService
}

// NewModuleItemHandler crea una nueva instancia
func NewModuleItemHandler(db *gorm.DB) *ModuleItemHandler {
  return &ModuleItemHandler{
    service: NewModuleItemService(db), // ← Inyectar servicio
  }
}

// Index godoc
// @Summary Listar items de un módulo de perfil
// @Description Retorna los items/permisos asignados a un módulo específico de un perfil
// @Tags Profile Module Items
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param moduleId path uint true "ID del módulo"
// @Success 200 {array} dto.ItemResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items [get]
func (h *ModuleItemHandler) Index(c *gin.Context) {
  profileID, ok := common.ParseID(c, "id")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de perfil inválido",
    })
    return
  }
  moduleID, ok := common.ParseID(c, "moduleId")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de módulo inválido",
    })
    return
  }

  items, err := h.service.ListByModuleAndProfile(uint(profileID), uint(moduleID))
  if err != nil {
    if err.Error() == "el módulo no está asignado a este perfil" {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": "El módulo no está asignado a este perfil",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar items",
    })
    return
  }
  c.JSON(http.StatusOK, items)
}

// Sync godoc
// @Summary Sincronizar items de un módulo de perfil
// @Description Asigna o remueve items/permisos de un módulo específico (sync)
// @Tags Profile Module Items
// @Accept json
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param moduleId path uint true "ID del módulo"
// @Param items body dto.ProfileModuleItemSyncRequest true "Array de IDs de items"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items [put]
func (h *ModuleItemHandler) Sync(c *gin.Context) {
  profileID, ok := common.ParseID(c, "id")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de perfil inválido",
    })
    return
  }
  moduleID, ok := common.ParseID(c, "moduleId")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de módulo inválido",
    })
    return
  }

  var req dto.ProfileModuleItemSyncRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "Datos inválidos",
    })
    return
  }

  err := h.service.Sync(uint(profileID), uint(moduleID), req.Items)
  if err != nil {
    if err.Error() == "el módulo no está asignado a este perfil" {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": "El módulo no está asignado a este perfil",
      })
      return
    }
    if err.Error() == "uno o más items no existen" {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "Uno o más items no existen",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al sincronizar items",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Items actualizados exitosamente",
  })
}

// Attach godoc
// @Summary Agregar item individual a un módulo de perfil
// @Description Asigna un item/permiso específico sin afectar los demás
// @Tags Profile Module Items
// @Accept json
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param moduleId path uint true "ID del módulo"
// @Param item body dto.ProfileModuleItemAttachRequest true "ID del item"
// @Success 201 {object} dto.ItemResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items [post]
func (h *ModuleItemHandler) Attach(c *gin.Context) {
  profileID, ok := common.ParseID(c, "id")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de perfil inválido",
    })
    return
  }
  moduleID, ok := common.ParseID(c, "moduleId")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de módulo inválido",
    })
    return
  }

  var req dto.ProfileModuleItemAttachRequest
  if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "item_id es obligatorio",
    })
    return
  }

  item, err := h.service.Attach(uint(profileID), uint(moduleID), req.ItemID)
  if err != nil {
    if err.Error() == "el módulo no está asignado a este perfil" {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": "El módulo no está asignado a este perfil",
      })
      return
    }
    if err.Error() == "el item no existe" {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": "El item no existe",
      })
      return
    }
    if err.Error() == "este item ya está asignado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "Este item ya está asignado",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al asignar el item",
    })
    return
  }
  c.JSON(http.StatusCreated, item)
}

// Detach godoc
// @Summary Eliminar item de un módulo de perfil
// @Description Remueve un item/permiso específico de un módulo de perfil
// @Tags Profile Module Items
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param moduleId path uint true "ID del módulo"
// @Param itemId path uint true "ID del item"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items/{itemId} [delete]
func (h *ModuleItemHandler) Detach(c *gin.Context) {
  profileID, ok := common.ParseID(c, "id")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de perfil inválido",
    })
    return
  }
  moduleID, ok := common.ParseID(c, "moduleId")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de módulo inválido",
    })
    return
  }
  itemID, ok := common.ParseID(c, "itemId")
  if !ok {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "ID de item inválido",
    })
    return
  }

  err := h.service.Detach(uint(profileID), uint(moduleID), uint(itemID))
  if err != nil {
    if err.Error() == "el módulo no está asignado a este perfil" {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": "El módulo no está asignado a este perfil",
      })
      return
    }
    if err.Error() == "el item no está asignado" {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": "El item no está asignado",
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