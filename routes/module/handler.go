package module

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

// Handler maneja las operaciones HTTP para Modules
type Handler struct {
	db *gorm.DB
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Index godoc
// @Summary Listar módulos
// @Description Retorna todos los módulos registrados ordenados por ID descendente
// @Tags Modules
// @Produce json
// @Success 200 {array} dto.ModuleResponse
// @Router /modules [get]
func (h *Handler) Index(c *gin.Context) {
	var modules []model.Module
	if err := h.db.Order("id desc").Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar módulos",
		})
		return
	}

	response := make(dto.ModulesResponse, len(modules))
	for i, m := range modules {
		response[i] = dto.ModuleResponse{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
		}
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

	var module model.Module
	if err := h.db.First(&module, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	c.JSON(http.StatusOK, dto.ModuleResponse{
		ID:          module.ID,
		Name:        module.Name,
		Description: module.Description,
	})
}

// Create godoc
// @Summary Crear nuevo módulo
// @Description Crea un nuevo módulo con nombre y descripción
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

	// Validar que no exista el nombre (case-insensitive)
	var existing model.Module
	if err := h.db.Where("LOWER(name) = LOWER(?)", body.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	newModule := model.Module{
		Name:        body.Name,
		Description: body.Description,
	}
	if err := h.db.Create(&newModule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al crear el módulo",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ModuleResponse{
		ID:          newModule.ID,
		Name:        newModule.Name,
		Description: newModule.Description,
	})
}

// Update godoc
// @Summary Actualizar módulo
// @Description Actualiza nombre y descripción de un módulo existente
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

	var module model.Module
	if err := h.db.First(&module, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// Validar nombre único (excluyendo el propio registro)
	var existing model.Module
	if err := h.db.Where("LOWER(name) = LOWER(?) AND id != ?", body.Name, id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	module.Name = body.Name
	module.Description = body.Description
	if err := h.db.Save(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al actualizar el módulo",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ModuleResponse{
		ID:          module.ID,
		Name:        module.Name,
		Description: module.Description,
	})
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

	var module model.Module
	if err := h.db.First(&module, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// 🔒 Validar dependencias: ¿está asignado a algún profile_module?
	var count int64
	h.db.Model(&model.ProfileModule{}).Where("module_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrHasDependencies.Error(),
		})
		return
	}

	if err := h.db.Delete(&module).Error; err != nil {
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