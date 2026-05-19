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
// @Summary Listar módulos con paginación y búsqueda
// @Description Retorna módulos paginados, con búsqueda por nombre y ordenamiento
// @Tags Modules
// @Produce json
// @Param page query int false "Página" default(1)
// @Param per_page query int false "Registros por página" default(20)
// @Param search query string false "Término de búsqueda"
// @Param field query string false "Campo a buscar" default(name) Enums(name)
// @Param sort_by query string false "Campo para ordenar" Enums(id,name,slug,created_at)
// @Param sort_dir query string false "Dirección" Enums(asc,desc)
// @Success 200 {object} common.PaginatedResponse[dto.ModuleResponse]
// @Router /modules [get]
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
	allowedSortFields := []string{"id", "name", "slug", "created_at"}

	// Construir query base
	query := h.db.Model(&model.Module{})

	// Aplicar búsqueda y ordenamiento
	query = params.Apply(query, searchableFields, allowedSortFields)

	// Obtener total para metadatos de paginación
	var total int64
	h.db.Model(&model.Module{}).Count(&total)

	// Aplicar paginación y ejecutar
	offset := (params.Page - 1) * params.PerPage
	var modules []model.Module
	if err := query.Offset(offset).Limit(params.PerPage).Find(&modules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar módulos",
		})
		return
	}

	// Mapear a DTOs
	response := make([]dto.ModuleResponse, len(modules))
	for i, m := range modules {
		response[i] = dto.ModuleResponse{
			ID:   m.ID,
			Name: m.Name,
			Slug: m.Slug,
		}
	}

	// Retornar respuesta compatible con React PaginationInfo
	c.JSON(http.StatusOK, common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage))
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
		ID:   module.ID,
		Name: module.Name,
		Slug: module.Slug,
	})
}

// Create godoc
// @Summary Crear nuevo módulo
// @Description Crea un nuevo módulo con nombre y path (slug)
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

	// Validar que no exista el slug/path
	if err := h.db.Where("slug = ?", body.Slug).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "El path ya está en uso",
			"field":   "slug",
		})
		return
	}

	newModule := model.Module{
		Name: body.Name,
		Slug: body.Slug,
	}
	if err := h.db.Create(&newModule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al crear el módulo",
		})
		return
	}

	c.JSON(http.StatusCreated, dto.ModuleResponse{
		ID:   newModule.ID,
		Name: newModule.Name,
		Slug: newModule.Slug,
	})
}

// Update godoc
// @Summary Actualizar módulo
// @Description Actualiza nombre y path de un módulo existente
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

	// Validar slug único (excluyendo el propio registro)
	if err := h.db.Where("slug = ? AND id != ?", body.Slug, id).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": "El path ya está en uso",
			"field":   "slug",
		})
		return
	}

	module.Name = body.Name
	module.Slug = body.Slug
	if err := h.db.Save(&module).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al actualizar el módulo",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ModuleResponse{
		ID:   module.ID,
		Name: module.Name,
		Slug: module.Slug,
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