package state

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

// Handler maneja las operaciones HTTP para States
type Handler struct {
	db *gorm.DB
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Index godoc
// @Summary Listar estados
// @Description Retorna todos los estados registrados ordenados por ID descendente
// @Tags States
// @Produce json
// @Success 200 {array} dto.StateResponse
// @Router /states [get]
func (h *Handler) Index(c *gin.Context) {
	var states []model.State
	if err := h.db.Order("id desc").Find(&states).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar estados",
		})
		return
	}

	response := make(dto.StatesResponse, len(states))
	for i, s := range states {
		response[i] = dto.StateResponse{ID: s.ID, Name: s.Name}
	}
	c.JSON(http.StatusOK, response)
}

// Show godoc
// @Summary Obtener estado por ID
func (h *Handler) Show(c *gin.Context) {
	id, ok := common.ParseID(c, "id")
	if !ok { return }

	var state model.State
	if err := h.db.First(&state, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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
	c.JSON(http.StatusOK, dto.StateResponse{ID: state.ID, Name: state.Name})
}

// Create godoc
// @Summary Crear nuevo estado
func (h *Handler) Create(c *gin.Context) {
	body, ok := common.BindAndValidate[dto.StateCreateRequest](c)
	if !ok { return }

	var existing model.State
	if err := h.db.Where("LOWER(name) = LOWER(?)", body.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	newState := model.State{Name: body.Name}
	if err := h.db.Create(&newState).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al crear el estado",
		})
		return
	}
	c.JSON(http.StatusCreated, dto.StateResponse{ID: newState.ID, Name: newState.Name})
}

// Update godoc
// @Summary Actualizar estado
func (h *Handler) Update(c *gin.Context) {
	id, ok := common.ParseID(c, "id")
	if !ok { return }

	body, ok := common.BindAndValidate[dto.StateUpdateRequest](c)
	if !ok { return }

	var state model.State
	if err := h.db.First(&state, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	var existing model.State
	if err := h.db.Where("LOWER(name) = LOWER(?) AND id != ?", body.Name, id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	state.Name = body.Name
	if err := h.db.Save(&state).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al actualizar el estado",
		})
		return
	}
	c.JSON(http.StatusOK, dto.StateResponse{ID: state.ID, Name: state.Name})
}

// Delete godoc
// @Summary Eliminar estado
func (h *Handler) Delete(c *gin.Context) {
	id, ok := common.ParseID(c, "id")
	if !ok { return }

	var state model.State
	if err := h.db.First(&state, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	var count int64
	h.db.Model(&model.UserMetadata{}).Where("state = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrHasDependencies.Error(),
		})
		return
	}

	if err := h.db.Delete(&state).Error; err != nil {
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