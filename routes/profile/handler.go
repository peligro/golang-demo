package profile

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
)

// Handler maneja las operaciones HTTP para Profiles
type Handler struct {
	db *gorm.DB
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Index godoc
// @Summary Listar perfiles
// @Description Retorna todos los perfiles registrados ordenados por ID descendente
// @Tags Profiles
// @Produce json
// @Success 200 {array} dto.ProfileResponse
// @Router /profiles [get]
func (h *Handler) Index(c *gin.Context) {
	var profiles []model.Profile
	if err := h.db.Order("id desc").Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar perfiles",
		})
		return
	}

	response := make(dto.ProfilesResponse, len(profiles))
	for i, p := range profiles {
		response[i] = dto.ProfileResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
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
// @Router /profiles/{id} [get]
func (h *Handler) Show(c *gin.Context) {
	id, ok := common.ParseID(c, "id")
	if !ok {
		return
	}

	var profile model.Profile
	if err := h.db.First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	c.JSON(http.StatusOK, dto.ProfileResponse{
		ID:          profile.ID,
		Name:        profile.Name,
		Description: profile.Description,
	})
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
// @Router /profiles [post]
func (h *Handler) Create(c *gin.Context) {
	body, ok := common.BindAndValidate[dto.ProfileCreateRequest](c)
	if !ok {
		return
	}

	// Validar que no exista el nombre (case-insensitive)
	var existing model.Profile
	if err := h.db.Where("LOWER(name) = LOWER(?)", body.Name).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	newProfile := model.Profile{
		Name:        body.Name,
		Description: body.Description,
	}
	if err := h.db.Create(&newProfile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al crear el perfil",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ProfileResponse{
		ID:          newProfile.ID,
		Name:        newProfile.Name,
		Description: newProfile.Description,
	})
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

	var profile model.Profile
	if err := h.db.First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// Validar nombre único (excluyendo el propio registro)
	var existing model.Profile
	if err := h.db.Where("LOWER(name) = LOWER(?) AND id != ?", body.Name, id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrDuplicate.Error(),
			"field":   "name",
		})
		return
	}

	profile.Name = body.Name
	profile.Description = body.Description
	if err := h.db.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al actualizar el perfil",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ProfileResponse{
		ID:          profile.ID,
		Name:        profile.Name,
		Description: profile.Description,
	})
}

// Delete godoc
// @Summary Eliminar perfil
// @Description Elimina un perfil por su ID (valida dependencias con profile_module)
// @Tags Profiles
// @Produce json
// @Param id path uint true "ID del perfil"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /profiles/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, ok := common.ParseID(c, "id")
	if !ok {
		return
	}

	var profile model.Profile
	if err := h.db.First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
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

	// 🔒 Validar dependencias: ¿está asignado a algún profile_module?
	var count int64
	h.db.Model(&model.ProfileModule{}).Where("profile_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": common.ErrHasDependencies.Error(),
		})
		return
	}

	if err := h.db.Delete(&profile).Error; err != nil {
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