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
// @Router /profiles [get]
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

	// Whitelist de campos permitidos
	searchableFields := []string{"name"}
	allowedSortFields := []string{"id", "name", "created_at"}

	// Construir query base
	query := h.db.Model(&model.Profile{})

	// Aplicar búsqueda y ordenamiento
	query = params.Apply(query, searchableFields, allowedSortFields)

	// Obtener total para metadatos de paginación
	var total int64
	h.db.Model(&model.Profile{}).Count(&total)

	// Aplicar paginación y ejecutar
	offset := (params.Page - 1) * params.PerPage
	var profiles []model.Profile
	if err := query.Offset(offset).Limit(params.PerPage).Find(&profiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar perfiles",
		})
		return
	}

	// Mapear a DTOs
	response := make([]dto.ProfileResponse, len(profiles))
	for i, p := range profiles {
		response[i] = dto.ProfileResponse{
			ID:          p.ID,
			Name:        p.Name,
			Description: p.Description,
		}
	}

	// Retornar respuesta compatible con React PaginationInfo
	c.JSON(http.StatusOK, common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage))
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