package app_menu

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
// @Summary Listar todos los menús (sin paginación)
// @Description Retorna la lista completa de menús ordenada, ideal para el sidebar
// @Tags App Menu
// @Produce json
// @Success 200 {array} dto.AppMenuResponse
// @Failure 500 {object} map[string]string
// @Router /app-menu-all [get]
func (h *Handler) ListAll(c *gin.Context) {
  menus, err := h.service.ListAll()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar menús"})
    return
  }
  c.JSON(http.StatusOK, menus)
}

// List godoc
// @Summary Listar menús con paginación
// @Description Retorna menús paginados para el panel de administración
// @Tags App Menu
// @Produce json
// @Param page query int false "Página" default(1)
// @Param per_page query int false "Registros por página" default(20)
// @Param search query string false "Término de búsqueda"
// @Param field query string false "Campo a buscar" Enums(label,title)
// @Param sort_by query string false "Campo para ordenar" Enums(id,label,title,order,created_at)
// @Param sort_dir query string false "Dirección" Enums(asc,desc)
// @Success 200 {object} common.PaginatedResponse[dto.AppMenuResponse]
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /app-menu [get]
func (h *Handler) List(c *gin.Context) {
  params := common.DefaultPagination()
  if err := c.ShouldBindQuery(&params); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Parámetros inválidos"})
    return
  }

  response, err := h.service.ListPaginated(params)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al consultar menús"})
    return
  }
  c.JSON(http.StatusOK, response)
}

// Get godoc
// @Summary Obtener menú por ID
// @Description Retorna un menú específico
// @Tags App Menu
// @Produce json
// @Param id path uint true "ID del menú"
// @Success 200 {object} dto.AppMenuResponse
// @Failure 404 {object} map[string]string
// @Router /app-menu/{id} [get]
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
// @Summary Crear nuevo menú
// @Description Crea un menú con validación de padre y módulo
// @Tags App Menu
// @Accept json
// @Produce json
// @Param menu body dto.AppMenuCreateRequest true "Datos del menú"
// @Success 201 {object} dto.AppMenuResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /app-menu [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.AppMenuCreateRequest](c)
  if !ok { return }

  menu, err := h.service.Create(body)
  if err != nil {
    if err.Error() == "el menú padre no existe" || err.Error() == "el módulo no existe" {
      c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Error al crear"})
    return
  }
  c.JSON(http.StatusCreated, menu)
}

// Update godoc
// @Summary Actualizar menú
// @Description Actualiza un menú existente
// @Tags App Menu
// @Accept json
// @Produce json
// @Param id path uint true "ID del menú"
// @Param menu body dto.AppMenuUpdateRequest true "Nuevos datos"
// @Success 200 {object} dto.AppMenuResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /app-menu/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok { return }

  body, ok := common.BindAndValidate[dto.AppMenuUpdateRequest](c)
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
// @Summary Eliminar menú
// @Description Elimina un menú si no tiene hijos
// @Tags App Menu
// @Produce json
// @Param id path uint true "ID del menú"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /app-menu/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
  id, ok := common.ParseID(c, "id")
  if !ok { return }

  err := h.service.Delete(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": common.ErrNotFound.Error()})
      return
    }
    c.JSON(http.StatusConflict, gin.H{"status": "error", "message": err.Error()})
    return
  }
  c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Menú eliminado exitosamente"})
}
