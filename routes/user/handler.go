package user

import (
  "errors"
  "net/http"
  "strconv"

  "github.com/gin-gonic/gin"
  "gorm.io/gorm"

  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
)

// Handler maneja las operaciones HTTP para Users
type Handler struct {
  service *Service
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
  return &Handler{
    service: NewService(db),
  }
}

// Index godoc
// @Summary Listar usuarios con paginación y filtros avanzados
// @Description Retorna usuarios paginados con búsqueda por nombre/email, filtros por estado y perfil, y ordenamiento flexible
// @Tags Users
// @Produce json
// @Param page query int false "Página" default(1) minimum(1)
// @Param per_page query int false "Registros por página" default(20) minimum(1) maximum(100)
// @Param search query string false "Término de búsqueda"
// @Param field query string false "Campo a buscar" Enums(name,email) default(name)
// @Param state query int false "Filtrar por estado" Enums(0,1)
// @Param profile_id query uint false "Filtrar por perfil"
// @Param sort_by query string false "Campo para ordenar" Enums(id,name,email,created_at) default(id)
// @Param sort_dir query string false "Dirección de ordenamiento" Enums(asc,desc) default(desc)
// @Success 200 {object} common.PaginatedResponse[dto.UserResponse]
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [get]
func (h *Handler) Index(c *gin.Context) {
  // Parsear parámetros con valores por defecto
  params := userListParams{
    Page:    1,
    PerPage: 20,
    SortBy:  "id",
    SortDir: "desc",
  }

  if page := c.Query("page"); page != "" {
    p, err := strconv.Atoi(page)
    if err != nil || p < 1 {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "parámetro 'page' inválido",
      })
      return
    }
    params.Page = p
  }
  if perPage := c.Query("per_page"); perPage != "" {
    pp, err := strconv.Atoi(perPage)
    if err != nil || pp < 1 || pp > 100 {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "parámetro 'per_page' inválido (1-100)",
      })
      return
    }
    params.PerPage = pp
  }

  params.Search = c.Query("search")
  params.Field = c.Query("field")
  if params.Search != "" && params.Field != "" && params.Field != "name" && params.Field != "email" {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": "campo de búsqueda no permitido",
    })
    return
  }

  if stateStr := c.Query("state"); stateStr != "" {
    state, err := strconv.Atoi(stateStr)
    if err != nil || (state != 0 && state != 1) {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "parámetro 'state' inválido (0 o 1)",
      })
      return
    }
    params.State = &state
  }

  if profileIDStr := c.Query("profile_id"); profileIDStr != "" {
    profileID, err := strconv.ParseUint(profileIDStr, 10, 32)
    if err != nil {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "parámetro 'profile_id' inválido",
      })
      return
    }
    pid := uint(profileID)
    params.ProfileID = &pid
  }

  if sortBy := c.Query("sort_by"); sortBy != "" {
    params.SortBy = sortBy
  }
  if sortDir := c.Query("sort_dir"); sortDir != "" {
    params.SortDir = sortDir
  }

  response, err := h.service.ListPaginated(params)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar usuarios",
    })
    return
  }
  c.JSON(http.StatusOK, response)
}

// Show godoc
// @Summary Obtener usuario por ID
// @Description Retorna un usuario específico con su metadata y perfil asociado
// @Tags Users
// @Produce json
// @Param id path uint true "ID del usuario" minimum(1)
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [get]
func (h *Handler) Show(c *gin.Context) {
  id, err := parseUserID(c)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": err.Error(),
    })
    return
  }

  user, err := h.service.GetByID(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar el usuario",
    })
    return
  }
  c.JSON(http.StatusOK, user)
}

// Create godoc
// @Summary Crear nuevo usuario
// @Description Crea un usuario con validación de email único, password hasheado con bcrypt, y metadata opcional
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.UserCreateRequest true "Datos del usuario"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users [post]
func (h *Handler) Create(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.UserCreateRequest](c)
  if !ok {
    return
  }

  user, err := h.service.Create(
    body.Name,
    body.Email,
    body.Password,
    body.Phone,
    body.State,
    body.ProfileID,
  )
  if err != nil {
    if err.Error() == "email duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "El email ya está registrado",
        "field":   "email",
      })
      return
    }
    if err.Error() == "perfil no existe" {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "El perfil especificado no existe",
        "field":   "profile_id",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al crear el usuario",
    })
    return
  }
  c.JSON(http.StatusCreated, user)
}

// Update godoc
// @Summary Actualizar usuario
// @Description Actualiza datos del usuario, metadata y password (opcional) con validaciones robustas
// @Tags Users
// @Accept json
// @Produce json
// @Param id path uint true "ID del usuario" minimum(1)
// @Param user body dto.UserUpdateRequest true "Nuevos datos del usuario"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [put]
func (h *Handler) Update(c *gin.Context) {
  id, err := parseUserID(c)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": err.Error(),
    })
    return
  }

  body, ok := common.BindAndValidate[dto.UserUpdateRequest](c)
  if !ok {
    return
  }

  user, err := h.service.Update(
    uint(id),
    body.Name,
    body.Email,
    body.Password,
    body.Phone,
    body.State,
    body.ProfileID,
  )
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    if err.Error() == "email duplicado" {
      c.JSON(http.StatusConflict, gin.H{
        "status":  "error",
        "message": "El email ya está registrado",
        "field":   "email",
      })
      return
    }
    if err.Error() == "perfil no existe" {
      c.JSON(http.StatusBadRequest, gin.H{
        "status":  "error",
        "message": "El perfil especificado no existe",
        "field":   "profile_id",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al actualizar el usuario",
    })
    return
  }
  c.JSON(http.StatusOK, user)
}

// Delete godoc
// @Summary Eliminar usuario
// @Description Elimina un usuario y su metadata asociada de forma segura
// @Tags Users
// @Produce json
// @Param id path uint true "ID del usuario" minimum(1)
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /users/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
  id, err := parseUserID(c)
  if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "status":  "error",
      "message": err.Error(),
    })
    return
  }

  err = h.service.Delete(uint(id))
  if err != nil {
    if errors.Is(err, common.ErrNotFound) {
      c.JSON(http.StatusNotFound, gin.H{
        "status":  "error",
        "message": common.ErrNotFound.Error(),
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al eliminar el usuario",
    })
    return
  }
  c.JSON(http.StatusOK, gin.H{
    "status":  "ok",
    "message": "Usuario eliminado exitosamente",
  })
}

// Me godoc
// @Summary Obtener usuario autenticado
// @Description Retorna los datos del usuario autenticado (requiere middleware de auth)
// @Tags Users
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Router /users/me [get]
func (h *Handler) Me(c *gin.Context) {
  c.JSON(http.StatusUnauthorized, gin.H{
    "status":  "error",
    "message": "Requiere autenticación",
  })
}

// parseUserID valida y convierte un param "id" a uint
func parseUserID(c *gin.Context) (uint, error) {
  idStr := c.Param("id")
  id, err := strconv.ParseUint(idStr, 10, 32)
  if err != nil || id == 0 {
    return 0, errors.New("ID de usuario inválido")
  }
  return uint(id), nil
}