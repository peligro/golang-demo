package profile

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

// ModuleItemHandler maneja las operaciones para ProfileModuleItem (items/permisos)
type ModuleItemHandler struct {
	db *gorm.DB
}

// NewModuleItemHandler crea una nueva instancia
func NewModuleItemHandler(db *gorm.DB) *ModuleItemHandler {
	return &ModuleItemHandler{db: db}
}

// Index godoc
// @Summary Listar items de un módulo de perfil
// @Description Retorna los items/permisos asignados a un módulo específico de un perfil
// @Tags Profile Module Items
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param moduleId path uint true "ID del módulo"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items [get]
func (h *ModuleItemHandler) Index(c *gin.Context) {
	profileID, ok := common.ParseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de perfil inválido"})
		return
	}
	moduleID, ok := common.ParseID(c, "moduleId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de módulo inválido"})
		return
	}

	var profile model.Profile
	if err := h.db.First(&profile, profileID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": common.ErrNotFound.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	var profileModule model.ProfileModule
	if err := h.db.Preload("Module").Where("profile_id = ? AND module_id = ?", profileID, moduleID).First(&profileModule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "El módulo no está asignado a este perfil"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	var profileModuleItems []model.ProfileModuleItem
	if err := h.db.Where("profile_module_id = ?", profileModule.ID).Preload("Item").Find(&profileModuleItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar items"})
		return
	}

	items := make([]map[string]interface{}, len(profileModuleItems))
	itemIDs := make([]uint, len(profileModuleItems))
	for i, pmi := range profileModuleItems {
		itemName := ""
		if pmi.Item != nil {
			itemName = pmi.Item.Name
		}
		items[i] = map[string]interface{}{"id": pmi.ItemID, "name": itemName}
		itemIDs[i] = pmi.ItemID
	}

	moduleName := ""
	moduleSlug := ""
	if profileModule.Module != nil {
		moduleName = profileModule.Module.Name
		moduleSlug = profileModule.Module.Slug
	}

	c.JSON(http.StatusOK, gin.H{
		"profile_id": profile.ID, "profile_name": profile.Name,
		"module_id": moduleID, "module_name": moduleName, "module_slug": moduleSlug,
		"items": items, "item_ids": itemIDs, "total": len(items),
	})
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
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items [put]
func (h *ModuleItemHandler) Sync(c *gin.Context) {
	profileID, ok := common.ParseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de perfil inválido"})
		return
	}
	moduleID, ok := common.ParseID(c, "moduleId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de módulo inválido"})
		return
	}

	var req dto.ProfileModuleItemSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Datos inválidos"})
		return
	}

	var profileModule model.ProfileModule
	if err := h.db.Preload("Module").Where("profile_id = ? AND module_id = ?", profileID, moduleID).First(&profileModule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "El módulo no está asignado a este perfil"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	if len(req.Items) > 0 {
		var count int64
		h.db.Model(&model.Item{}).Where("id IN ?", req.Items).Count(&count)
		if count != int64(len(req.Items)) {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Uno o más items no existen"})
			return
		}
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("profile_module_id = ?", profileModule.ID).Delete(&model.ProfileModuleItem{}).Error; err != nil {
			return err
		}
		if len(req.Items) > 0 {
			pmis := make([]model.ProfileModuleItem, len(req.Items))
			for i, itemID := range req.Items {
				pmis[i] = model.ProfileModuleItem{ProfileModuleID: profileModule.ID, ItemID: itemID}
			}
			if err := tx.Create(&pmis).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al sincronizar items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok", "message": "Items actualizados exitosamente",
		"profile_id": profileID, "module_id": moduleID, "attached": len(req.Items), "items": req.Items,
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
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /profiles/{id}/modules/{moduleId}/items [post]
func (h *ModuleItemHandler) Attach(c *gin.Context) {
	profileID, ok := common.ParseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de perfil inválido"})
		return
	}
	moduleID, ok := common.ParseID(c, "moduleId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de módulo inválido"})
		return
	}

	var req dto.ProfileModuleItemAttachRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "item_id es obligatorio"})
		return
	}

	var profileModule model.ProfileModule
	if err := h.db.Where("profile_id = ? AND module_id = ?", profileID, moduleID).First(&profileModule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "El módulo no está asignado a este perfil"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	var item model.Item
	if err := h.db.First(&item, req.ItemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "El item no existe"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	var existing model.ProfileModuleItem
	if err := h.db.Where("profile_module_id = ? AND item_id = ?", profileModule.ID, req.ItemID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"status": "error", "message": "Este item ya está asignado"})
		return
	}

	pmi := model.ProfileModuleItem{ProfileModuleID: profileModule.ID, ItemID: req.ItemID}
	if err := h.db.Create(&pmi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al asignar el item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "ok", "message": "Item agregado exitosamente",
		"item": map[string]interface{}{"id": item.ID, "name": item.Name},
	})
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
// @Router /profiles/{id}/modules/{moduleId}/items/{itemId} [delete]
func (h *ModuleItemHandler) Detach(c *gin.Context) {
	profileID, ok := common.ParseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de perfil inválido"})
		return
	}
	moduleID, ok := common.ParseID(c, "moduleId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de módulo inválido"})
		return
	}
	itemID, ok := common.ParseID(c, "itemId")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de item inválido"})
		return
	}

	var profileModule model.ProfileModule
	if err := h.db.Where("profile_id = ? AND module_id = ?", profileID, moduleID).First(&profileModule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "El módulo no está asignado a este perfil"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	var pmi model.ProfileModuleItem
	if err := h.db.Where("profile_module_id = ? AND item_id = ?", profileModule.ID, itemID).First(&pmi).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "El item no está asignado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
		return
	}

	if err := h.db.Delete(&pmi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al eliminar el item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Item eliminado exitosamente"})
}