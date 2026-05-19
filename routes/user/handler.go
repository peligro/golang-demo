package user

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/peligro/golang-demo/common"
	"github.com/peligro/golang-demo/dto"
	"github.com/peligro/golang-demo/model"
	"github.com/peligro/golang-demo/pkg/auth"
)

// Handler maneja las operaciones HTTP para Users con arquitectura senior
type Handler struct {
	db *gorm.DB
}

// NewHandler inyecta dependencia de DB (patrón recomendado)
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// =============================================================================
// INDEX - Listado con paginación, búsqueda y filtros avanzados
// =============================================================================

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
	params, err := parseUserListParams(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Query base: usar comillas dobles para escapar "user" (keyword reservada en PostgreSQL)
	query := h.db.Table("\"user\"").
		Joins("LEFT JOIN user_metadata ON user_metadata.user_id = \"user\".id")

	// Aplicar búsqueda textual (name/email)
	if params.Search != "" && params.Field != "" {
		query = query.Where(fmt.Sprintf("\"user\".%s ILIKE ?", params.Field), "%"+params.Search+"%")
	}

	// Aplicar filtro por estado
	if params.State != nil {
		query = query.Where("user_metadata.state = ?", *params.State)
	}

	// Aplicar filtro por perfil
	if params.ProfileID != nil {
		query = query.Where("user_metadata.profile_id = ?", *params.ProfileID)
	}

	// Aplicar ordenamiento
	if isValidSortField(params.SortBy) {
		dir := strings.ToUpper(params.SortDir)
		if dir != "ASC" && dir != "DESC" {
			dir = "DESC"
		}
		query = query.Order(fmt.Sprintf("\"user\".%s %s", params.SortBy, dir))
	} else {
		query = query.Order("\"user\".id DESC")
	}

	// Obtener total para paginación
	var total int64
	if err := h.db.Table("\"user\"").Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al contar usuarios",
		})
		return
	}

	// Ejecutar consulta
	offset := (params.Page - 1) * params.PerPage
	var users []model.User
	if err := query.Offset(offset).Limit(params.PerPage).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al consultar usuarios",
		})
		return
	}

	// Mapear a DTOs
	response, err := mapUsersToResponse(h.db, users)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al procesar respuesta",
		})
		return
	}

	c.JSON(http.StatusOK, common.NewPaginatedResponse(response, int(total), params.Page, params.PerPage))
}

// =============================================================================
// SHOW - Obtener usuario por ID con metadata completa
// =============================================================================

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

	// Consultar usuario con preload de metadata y perfil
	var user model.User
	if err := h.db.Preload("Metadata.Profile").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	// Mapear a respuesta (con perfil anidado si existe)
	response, err := mapUserToResponse(h.db, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al procesar respuesta",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// =============================================================================
// CREATE - Crear usuario con validaciones robustas y bcrypt
// =============================================================================

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
	body, err := parseUserCreateRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Validar unicidad de email
	if err := h.validateEmailUnique(0, body.Email); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"status":  "error",
			"message": err.Error(),
			"field":   "email",
		})
		return
	}

	// Validar que el profile_id exista si se proporciona
	if body.ProfileID != 0 {
		if err := h.validateProfileExists(body.ProfileID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
				"field":   "profile_id",
			})
			return
		}
	}

	// Hashear password con bcrypt
	hashedPassword, err := auth.HashPassword(body.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al procesar la contraseña",
		})
		return
	}

	// ✅ State: calcular valor por defecto ANTES de la transacción (para usar en respuesta)
	stateValue := 1 // default: activo
	if body.State != nil {
		stateValue = *body.State
	}

	// Ejecutar en transacción para consistencia
	var newUser model.User
	var newMeta model.UserMetadata

	err = h.db.Transaction(func(tx *gorm.DB) error {
		newUser = model.User{
			Name:     body.Name,
			Email:    body.Email,
			Password: hashedPassword,
		}
		if err := tx.Create(&newUser).Error; err != nil {
			return fmt.Errorf("error al crear usuario: %w", err)
		}

		// Crear metadata si hay datos relevantes
		if body.Phone != "" || body.State != nil || body.ProfileID != 0 {
			newMeta = model.UserMetadata{
				UserID:    newUser.ID,
				Phone:     body.Phone,
				State:     stateValue,  // ← Usar variable del scope externo
				ProfileID: body.ProfileID,
			}
			if err := tx.Create(&newMeta).Error; err != nil {
				return fmt.Errorf("error al guardar metadata: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al crear el usuario",
		})
		return
	}

	// Obtener perfil para la respuesta
	var profileSummary *dto.ProfileSummary
	if body.ProfileID != 0 {
		var profile model.Profile
		if err := h.db.First(&profile, body.ProfileID).Error; err == nil {
			profileSummary = &dto.ProfileSummary{
				ID:   profile.ID,
				Name: profile.Name,
			}
		}
	}

	// Retornar respuesta (SIN password, con date/time formateados)
	c.JSON(http.StatusCreated, dto.UserResponse{
		ID:        newUser.ID,
		Name:      newUser.Name,
		Email:     newUser.Email,
		Date:      common.FormatDate(newUser.CreatedAt),
		Time:      common.FormatTime(newUser.CreatedAt),
		Phone:     body.Phone,
		State:     stateValue,  // ← Ahora sí está en scope ✅
		ProfileID: body.ProfileID,
		Profile:   profileSummary,
	})
}

// =============================================================================
// UPDATE - Actualizar usuario con validaciones y manejo seguro de password
// =============================================================================

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

	body, err := parseUserUpdateRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	// Buscar usuario existente
	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	// Validar unicidad de email
	if body.Email != "" && !strings.EqualFold(body.Email, user.Email) {
		if err := h.validateEmailUnique(user.ID, body.Email); err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"status":  "error",
				"message": err.Error(),
				"field":   "email",
			})
			return
		}
	}

	// Validar profile_id
	if body.ProfileID != 0 {
		if err := h.validateProfileExists(body.ProfileID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": err.Error(),
				"field":   "profile_id",
			})
			return
		}
	}

	// Ejecutar en transacción
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Actualizar campos del usuario
		updates := make(map[string]interface{})
		if body.Name != "" {
			updates["name"] = body.Name
		}
		if body.Email != "" {
			updates["email"] = body.Email
		}
		if body.Password != "" {
			hashed, err := auth.HashPassword(body.Password)
			if err != nil {
				return fmt.Errorf("error al hashear contraseña: %w", err)
			}
			updates["password"] = hashed
		}

		if len(updates) > 0 {
			if err := tx.Model(&user).Updates(updates).Error; err != nil {
				return fmt.Errorf("error al actualizar usuario: %w", err)
			}
		}

		// Actualizar o crear metadata (upsert)
		var meta model.UserMetadata
		if err := tx.Where("user_id = ?", user.ID).FirstOrCreate(&meta, model.UserMetadata{UserID: user.ID}).Error; err != nil {
			return fmt.Errorf("error al consultar metadata: %w", err)
		}

		metaUpdates := make(map[string]interface{})
		if body.Phone != "" {
			metaUpdates["phone"] = body.Phone
		}
		// State: actualizar solo si fue enviado explícitamente (incluso si es 0)
		if body.State != nil {
			metaUpdates["state"] = *body.State
		}
		// ProfileID: mismo patrón
		if c.Request.URL.Query().Get("profile_id") != "" || body.ProfileID != 0 || meta.ID == 0 {
			metaUpdates["profile_id"] = body.ProfileID
		}

		if len(metaUpdates) > 0 {
			if err := tx.Model(&meta).Updates(metaUpdates).Error; err != nil {
				return fmt.Errorf("error al actualizar metadata: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al actualizar el usuario",
		})
		return
	}

	// Recargar y retornar respuesta
	response, err := mapUserToResponse(h.db, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Error al procesar respuesta",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// =============================================================================
// DELETE - Eliminar usuario con cascade manual de metadata
// =============================================================================

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

	var user model.User
	if err := h.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", id).Delete(&model.UserMetadata{}).Error; err != nil {
			return fmt.Errorf("error al eliminar metadata: %w", err)
		}
		if err := tx.Delete(&user).Error; err != nil {
			return fmt.Errorf("error al eliminar usuario: %w", err)
		}
		return nil
	})

	if err != nil {
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

// =============================================================================
// ME - Endpoint placeholder para usuario autenticado
// =============================================================================

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

// =============================================================================
// HELPERS PRIVADOS
// =============================================================================

type userListParams struct {
	Page      int
	PerPage   int
	Search    string
	Field     string
	State     *int
	ProfileID *uint
	SortBy    string
	SortDir   string
}

func parseUserListParams(c *gin.Context) (*userListParams, error) {
	params := &userListParams{
		Page:    1,
		PerPage: 20,
		SortBy:  "id",
		SortDir: "desc",
	}

	if page := c.Query("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil || p < 1 {
			return nil, fmt.Errorf("parámetro 'page' inválido")
		}
		params.Page = p
	}
	if perPage := c.Query("per_page"); perPage != "" {
		pp, err := strconv.Atoi(perPage)
		if err != nil || pp < 1 || pp > 100 {
			return nil, fmt.Errorf("parámetro 'per_page' inválido (1-100)")
		}
		params.PerPage = pp
	}

	params.Search = c.Query("search")
	params.Field = c.Query("field")
	if params.Search != "" && params.Field != "" && params.Field != "name" && params.Field != "email" {
		return nil, fmt.Errorf("campo de búsqueda no permitido")
	}

	if stateStr := c.Query("state"); stateStr != "" {
		state, err := strconv.Atoi(stateStr)
		if err != nil || (state != 0 && state != 1) {
			return nil, fmt.Errorf("parámetro 'state' inválido (0 o 1)")
		}
		params.State = &state
	}

	if profileIDStr := c.Query("profile_id"); profileIDStr != "" {
		profileID, err := strconv.ParseUint(profileIDStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("parámetro 'profile_id' inválido")
		}
		pid := uint(profileID)
		params.ProfileID = &pid
	}

	if sortBy := c.Query("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}
	if sortDir := c.Query("sort_dir"); sortDir != "" {
		params.SortDir = strings.ToLower(sortDir)
	}

	return params, nil
}

func isValidSortField(field string) bool {
	allowed := map[string]bool{
		"id": true, "name": true, "email": true, "created_at": true,
	}
	return allowed[field]
}

func parseUserID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil || id == 0 {
		return 0, fmt.Errorf("ID de usuario inválido")
	}
	return uint(id), nil
}

func parseUserCreateRequest(c *gin.Context) (*dto.UserCreateRequest, error) {
	body := &dto.UserCreateRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		return nil, fmt.Errorf("datos inválidos: %v", err)
	}
	return body, nil
}

func parseUserUpdateRequest(c *gin.Context) (*dto.UserUpdateRequest, error) {
	body := &dto.UserUpdateRequest{}
	if err := c.ShouldBindJSON(body); err != nil {
		return nil, fmt.Errorf("datos inválidos: %v", err)
	}
	return body, nil
}

func (h *Handler) validateEmailUnique(excludeID uint, email string) error {
	query := h.db.Model(&model.User{}).Where("email ILIKE ?", email)
	if excludeID != 0 {
		query = query.Where("id != ?", excludeID)
	}
	var existing model.User
	if err := query.First(&existing).Error; err == nil {
		return fmt.Errorf("el email ya está registrado")
	}
	return nil
}

func (h *Handler) validateProfileExists(profileID uint) error {
	var profile model.Profile
	if err := h.db.First(&profile, profileID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("el perfil especificado no existe")
		}
		return fmt.Errorf("error al validar perfil")
	}
	return nil
}

func mapUsersToResponse(db *gorm.DB, users []model.User) ([]dto.UserResponse, error) {
	response := make([]dto.UserResponse, len(users))
	for i, u := range users {
		resp, err := mapUserToResponse(db, u)
		if err != nil {
			return nil, err
		}
		response[i] = resp
	}
	return response, nil
}

func mapUserToResponse(db *gorm.DB, user model.User) (dto.UserResponse, error) {
	var meta model.UserMetadata
	if err := db.Where("user_id = ?", user.ID).Preload("Profile").First(&meta).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return dto.UserResponse{}, fmt.Errorf("error al cargar metadata: %w", err)
	}

	var profileSummary *dto.ProfileSummary
	if meta.Profile != nil {
		profileSummary = &dto.ProfileSummary{
			ID:   meta.Profile.ID,
			Name: meta.Profile.Name,
		}
	}

	return dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Date:      common.FormatDate(user.CreatedAt),
		Time:      common.FormatTime(user.CreatedAt),
		Phone:     meta.Phone,
		State:     meta.State,
		ProfileID: meta.ProfileID,
		Profile:   profileSummary,
	}, nil
}