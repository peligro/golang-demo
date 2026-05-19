package item

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

// Handler maneja las operaciones HTTP para Items
type Handler struct {
	db *gorm.DB
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Index godoc
// @Summary Listar items
// @Description Retorna todos los items registrados ordenados por ID descendente
// @Tags Items
// @Produce json
// @Success 200 {array} dto.ItemResponse
// @Router /items [get]
func (h *Handler) Index(c *gin.Context) {
	var items []model.Item
	if err := h.db.Order("id desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar items",
		})
		return
	}

	response := make(dto.ItemsResponse, len(items))
	for i, it := range items {
		response[i] = dto.ItemResponse{ID: it.ID, Name: it.Name}
	}
	c.JSON(http.StatusOK, response)
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

	var item model.Item
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	c.JSON(http.StatusOK, dto.ItemResponse{ID: item.ID, Name: item.Name})
}

// Create godoc
// @Summary Crear nuevo item
// @Description Crea un nuevo item (acción) con nombre único
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

	// Validar que no exista el nombre (case-insensitive)
	var existing model.Item
	if err := h.db.Where("LOWER(name) = LOWER(?)", body.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	newItem := model.Item{Name: body.Name}
	if err := h.db.Create(&newItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al crear el item",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ItemResponse{ID: newItem.ID, Name: newItem.Name})
}

// Update godoc
// @Summary Actualizar item
// @Description Actualiza el nombre de un item existente
// @Tags Items
// @Accept json
// @Produce json
// @Param id path uint true "ID del item"
// @Param item body dto.ItemUpdateRequest true "Nuevo nombre"
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

	var item model.Item
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// Validar nombre único (excluyendo el propio registro)
	var existing model.Item
	if err := h.db.Where("LOWER(name) = LOWER(?) AND id != ?", body.Name, id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	item.Name = body.Name
	if err := h.db.Save(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al actualizar el item",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ItemResponse{ID: item.ID, Name: item.Name})
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

	var item model.Item
	if err := h.db.First(&item, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// 🔒 Validar dependencias: ¿está asignado a algún profile_module_item?
	var count int64
	h.db.Model(&model.ProfileModuleItem{}).Where("item_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrHasDependencies.Error(),
		})
		return
	}

	if err := h.db.Delete(&item).Error; err != nil {
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