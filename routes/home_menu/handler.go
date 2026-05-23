package home_menu

import (
  "errors"
  "net/http"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

type Handler struct {
  service *Service
}

func NewHandler(db *gorm.DB) *Handler {
  return &Handler{service: NewService(db)}
}

// ListAll godoc
// @Summary Listar todos los menús del home (sin paginación)
// @Description Retorna la lista completa de menús del home ordenada, ideal para mostrar en el dashboard
// @Tags Home Menu
// @Produce json
// @Success 200 {array} dto.HomeMenuResponse
// @Failure 500 {object} map[string]string
// @Router /home-menu-all [get]
func (h *Handler) ListAll(c *gin.Context) {
  menus, err := h.service.ListAll()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar menús del home"})
    return
  }
  c.JSON(http.StatusOK, menus)
}

// List godoc
// @Summary Listar menús del home con paginación
// @Description Retorna menús del home paginados para el panel de administración
// @Tags Home Menu
// @Produce json
// @Param page query int false "Página" default(1)
// @Param per_page query int false "Registros por página" default(20)
// @Param search query string false "Término de búsqueda"
// @Param field query string false "Campo a buscar" Enums(title,description,slug)
// @Param sort_by query string false "Campo para ordenar" Enums(id,title,order,created_at)
// @Param sort_dir query string false "Dirección" Enums(asc,desc)
// @Success 200 {object} common.PaginatedResponse[dto.HomeMenuResponse]
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /home-menu [get]
func (h *Handler) List(c *gin.Context) {
  params := common.DefaultPagination()
  if err := c.ShouldBindQuery(&params); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Parámetros inválidos"})
    return
  }

  response, err := h.service.ListPaginated(params)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar menús del home"})
    return
  }
  c.JSON(http.StatusOK, response)
}

// Get godoc
// @Summary Obtener menú del home por ID
// @Description Retorna un menú del home específico
// @Tags Home Menu
// @Produce json
// @Param id path uint true "ID del menú del home"
// @Success 200 {object} dto.HomeMenuResponse
// @Failure 404 {object} map[string]string
// @Router /home-menu/{id} [get]
func (h *Handler) Get(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok { return }

  menu, err := h.service.GetByID(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": common.ErrNotFound.Error()})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar"})
    return
  }
  c.JSON(http.StatusOK, menu)
}

// Create godoc
// @Summary Crear nuevo menú del home
// @Description Crea un menú del home con validación de módulo
// @Tags Home Menu
// @Accept json
// @Produce json
// @Param menu body dto.HomeMenuCreateRequest true "Datos del menú del home"
// @Success 201 {object} dto.HomeMenuResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /home-menu [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.HomeMenuCreateRequest](c)
  if !ok { return }

  menu, err := h.service.Create(body)
  if err != nil {
    if err.Error() == "el módulo no existe" {
      c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al crear"})
    return
  }
  c.JSON(http.StatusCreated, menu)
}

// Update godoc
// @Summary Actualizar menú del home
// @Description Actualiza un menú del home existente
// @Tags Home Menu
// @Accept json
// @Produce json
// @Param id path uint true "ID del menú del home"
// @Param menu body dto.HomeMenuUpdateRequest true "Nuevos datos"
// @Success 200 {object} dto.HomeMenuResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /home-menu/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok { return }

  body, ok := common.BindAndValidate[dto.HomeMenuUpdateRequest](c)
  if !ok { return }

  menu, err := h.service.Update(uint(id), body)
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": common.ErrNotFound.Error()})
      return
    }
    c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
    return
  }
  c.JSON(http.StatusOK, menu)
}

// Delete godoc
// @Summary Eliminar menú del home
// @Description Elimina un menú del home
// @Tags Home Menu
// @Produce json
// @Param id path uint true "ID del menú del home"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /home-menu/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok { return }

  err := h.service.Delete(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": common.ErrNotFound.Error()})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Menú del home eliminado exitosamente"})
}
