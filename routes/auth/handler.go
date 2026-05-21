package auth

import (
  "net/http"

  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/middleware"
  "github.com/peligro/golang-demo/model"
  
  // 👇 ALIAS para evitar conflicto con el paquete local "auth"
  authpkg "github.com/peligro/golang-demo/pkg/auth"
  
  "gorm.io/gorm"
)

// Handler maneja las operaciones de autenticación
type Handler struct {
  db *gorm.DB
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
  return &Handler{db: db}
}

// Login godoc
// @Summary Iniciar sesión
// @Description Autentica al usuario y crea una sesión con cookie HTTP-only
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Credenciales"
// @Success 200 {object} dto.LoginResponse
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
  // 🔍 Verificar si ya está autenticado
  token, err := c.Cookie("auth_token")
  if err == nil {
    // Si existe la cookie, verificar si la sesión es válida en Redis
    if session, err := authpkg.GetSession(token); err == nil {
      // ✅ Ya está logueado, retornar sus datos
      var user model.User
      if err := h.db.First(&user, session.UserID).Error; err == nil {
        var meta model.UserMetadata
        if err := h.db.Where("user_id = ?", user.ID).Preload("Profile").First(&meta).Error; err != nil && err != gorm.ErrRecordNotFound {
          c.JSON(http.StatusInternalServerError, gin.H{
            "status":  "error",
            "message": "Error al cargar datos del usuario",
          })
          return
        }
        
        var profileSummary *dto.ProfileSummary
        if meta.Profile != nil {
          profileSummary = &dto.ProfileSummary{
            ID:   meta.Profile.ID,
            Name: meta.Profile.Name,
          }
        }

        c.JSON(http.StatusOK, gin.H{
          "status":  "ok",
          "message": "Ya estás autenticado",
          "user": dto.UserResponse{
            ID:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
            Date:      common.FormatDate(user.CreatedAt),
            Time:      common.FormatTime(user.CreatedAt),
            Phone:     meta.Phone,
            State:     meta.State,
            ProfileID: meta.ProfileID,
            Profile:   profileSummary,
          },
        })
        return
      }
    }
  }

  // 🔐 Si no está autenticado, proceder con el login normal
  body, ok := common.BindAndValidate[dto.LoginRequest](c)
  if !ok {
    return
  }

  // 🔍 Buscar usuario CON su metadata para validar state
  var user model.User
  var meta model.UserMetadata
  
  if err := h.db.Where("email ILIKE ?", body.Email).First(&user).Error; err != nil {
    if err == gorm.ErrRecordNotFound {
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "Credenciales inválidas",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar usuario",
    })
    return
  }

  // 🔒 VALIDAR ESTADO ANTES DE VERIFICAR PASSWORD
  if err := h.db.Where("user_id = ?", user.ID).First(&meta).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al cargar datos del usuario",
    })
    return
  }

  // 🔒 BLOQUEAR SI EL USUARIO ESTÁ INACTIVO
  if meta.State != 1 {
    c.JSON(http.StatusForbidden, gin.H{
      "status":  "error",
      "message": "Usuario inactivo. Contacta al administrador",
      "state":   meta.State,
    })
    return
  }

  // Verificar contraseña (solo si está activo)
  if !user.CheckPassword(body.Password) {
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "Credenciales inválidas",
    })
    return
  }

  token, err = authpkg.GenerateToken()
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al generar token",
    })
    return
  }

  if err := authpkg.CreateSession(token, user.ID, user.Email); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al crear sesión",
    })
    return
  }

  middleware.SetAuthCookie(c, token)

  if err := h.db.Where("user_id = ?", user.ID).Preload("Profile").First(&meta).Error; err != nil && err != gorm.ErrRecordNotFound {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al cargar datos del usuario",
    })
    return
  }

  var profileSummary *dto.ProfileSummary
  if meta.Profile != nil {
    profileSummary = &dto.ProfileSummary{
      ID:   meta.Profile.ID,
      Name: meta.Profile.Name,
    }
  }

  response := dto.LoginResponse{
    User: dto.UserResponse{
      ID:        user.ID,
      Name:      user.Name,
      Email:     user.Email,
      Date:      common.FormatDate(user.CreatedAt),
      Time:      common.FormatTime(user.CreatedAt),
      Phone:     meta.Phone,
      State:     meta.State,
      ProfileID: meta.ProfileID,
      Profile:   profileSummary,
    },
  }

  c.JSON(http.StatusOK, response)
}
// Logout godoc
// @Summary Cerrar sesión
// @Description Invalida la sesión actual y elimina la cookie
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.LogoutResponse
// @Failure 401 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
  // 🔍 Obtener token de la cookie
  token, err := c.Cookie("auth_token")
  if err != nil {
    // ❌ No hay cookie → 401
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No hay sesión activa",
    })
    return
  }

  // 🔍 Verificar si la sesión existe en Redis
  _, err = authpkg.GetSession(token)
  if err != nil {
    // ❌ La sesión no existe o expiró → 401
    middleware.ClearAuthCookie(c) // Limpiar cookie inválida
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No hay sesión activa",
    })
    return
  }

  // ✅ Eliminar sesión de Redis
  _ = authpkg.DeleteSession(token)

  // ✅ Eliminar cookie
  middleware.ClearAuthCookie(c)

  c.JSON(http.StatusOK, dto.LogoutResponse{
    Message: "Sesión cerrada exitosamente",
  })
}

// Me godoc
// @Summary Obtener usuario autenticado
// @Description Retorna los datos del usuario actualmente autenticado
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.MeResponse
// @Failure 401 {object} map[string]string
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
  userID, exists := middleware.GetUserIDFromContext(c)
  if !exists {
    // Esto nunca debería pasar porque el middleware ya validó
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No autenticado",
    })
    return
  }

  var user model.User
  var meta model.UserMetadata
  
  if err := h.db.First(&user, userID).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al consultar usuario",
    })
    return
  }
  
  if err := h.db.Where("user_id = ?", userID).Preload("Profile").First(&meta).Error; err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al cargar datos del usuario",
    })
    return
  }

  var profileSummary *dto.ProfileSummary
  if meta.Profile != nil {
    profileSummary = &dto.ProfileSummary{
      ID:   meta.Profile.ID,
      Name: meta.Profile.Name,
    }
  }

  response := dto.MeResponse{
    User: dto.UserResponse{
      ID:        user.ID,
      Name:      user.Name,
      Email:     user.Email,
      Date:      common.FormatDate(user.CreatedAt),
      Time:      common.FormatTime(user.CreatedAt),
      Phone:     meta.Phone,
      State:     meta.State,
      ProfileID: meta.ProfileID,
      Profile:   profileSummary,
    },
  }

  c.JSON(http.StatusOK, response)
}

// Refresh godoc
// @Summary Renovar sesión
// @Description Extiende el TTL de la sesión actual
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.RefreshResponse
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
  token, err := c.Cookie("auth_token")
  if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No autenticado",
    })
    return
  }

  _, err = authpkg.GetSession(token)
  if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "Sesión inválida o expirada",
    })
    return
  }

  if err := authpkg.ExtendSession(token); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al renovar sesión",
    })
    return
  }

  c.JSON(http.StatusOK, dto.RefreshResponse{
    Message: "Sesión renovada exitosamente",
  })
}