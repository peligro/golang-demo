package profile

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

type ModuleHandler struct {
	db *gorm.DB
}

func NewModuleHandler(db *gorm.DB) *ModuleHandler {
	return &ModuleHandler{db: db}
}

// Index godoc
// @Summary Listar módulos asignados a un perfil
// @Description Retorna los módulos asignados a un perfil específico
// @Tags Profile Modules
// @Produce json
// @Param id path uint true "ID del perfil"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /profiles/{id}/modules [get]
func (h *ModuleHandler) Index(c *gin.Context) {
	profileID, ok := common.ParseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de perfil inválido"})
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

	var profileModules []model.ProfileModule
	if err := h.db.Where("profile_id = ?", profileID).Preload("Module").Find(&profileModules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar módulos"})
		return
	}

	modules := make([]map[string]interface{}, len(profileModules))
	moduleIDs := make([]uint, len(profileModules))
	for i, pm := range profileModules {
		modules[i] = map[string]interface{}{"id": pm.ModuleID, "name": "", "slug": ""}
		if pm.Module != nil {
			modules[i]["name"] = pm.Module.Name
			modules[i]["slug"] = pm.Module.Slug
		}
		moduleIDs[i] = pm.ModuleID
	}

	c.JSON(http.StatusOK, gin.H{
		"profile_id": profile.ID, "profile_name": profile.Name,
		"modules": modules, "module_ids": moduleIDs, "total": len(modules),
	})
}

// Sync godoc
// @Summary Sincronizar módulos de un perfil
// @Description Asigna o remueve módulos de un perfil (operación sync)
// @Tags Profile Modules
// @Accept json
// @Produce json
// @Param id path uint true "ID del perfil"
// @Param modules body dto.ProfileModuleSyncRequest true "Array de IDs de módulos"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /profiles/{id}/modules [put]
func (h *ModuleHandler) Sync(c *gin.Context) {
	profileID, ok := common.ParseID(c, "id")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "ID de perfil inválido"})
		return
	}

	var req dto.ProfileModuleSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Datos inválidos"})
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

	if len(req.Modules) > 0 {
		var count int64
		h.db.Model(&model.Module{}).Where("id IN ?", req.Modules).Count(&count)
		if count != int64(len(req.Modules)) {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Uno o más módulos no existen"})
			return
		}
	}

	// ✅ CORRECCIÓN: convertir profileID (uint64) a uint
	profileIDUint := uint(profileID)

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("profile_id = ?", profileIDUint).Delete(&model.ProfileModule{}).Error; err != nil {
			return err
		}
		if len(req.Modules) > 0 {
			pms := make([]model.ProfileModule, len(req.Modules))
			for i, moduleID := range req.Modules {
				pms[i] = model.ProfileModule{ProfileID: profileIDUint, ModuleID: moduleID}
			}
			if err := tx.Create(&pms).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al sincronizar módulos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok", "message": "Módulos actualizados exitosamente",
		"profile_id": profileID, "attached": len(req.Modules), "modules": req.Modules,
	})
}