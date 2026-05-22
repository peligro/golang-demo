package auth

import (
  "net/http"
  
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/dto"
  "github.com/peligro/golang-demo/middleware"
  authpkg "github.com/peligro/golang-demo/pkg/auth"
  "gorm.io/gorm"
)

// Handler maneja las operaciones HTTP para autenticación
type Handler struct {
  service *Service
}

// NewHandler crea una nueva instancia
func NewHandler(db *gorm.DB) *Handler {
  return &Handler{
    service: NewService(db),
  }
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
// @Failure 500 {object} map[string]string
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
  body, ok := common.BindAndValidate[dto.LoginRequest](c)
  if !ok {
    return
  }

  user, token, err := h.service.Login(body.Email, body.Password)
  if err != nil {
    if err.Error() == "credenciales inválidas" {
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "Credenciales inválidas",
      })
      return
    }
    if err.Error() == "usuario inactivo" {
      c.JSON(http.StatusForbidden, gin.H{
        "status":  "error",
        "message": "Usuario inactivo. Contacta al administrador",
      })
      return
    }
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al iniciar sesión",
    })
    return
  }

  middleware.SetAuthCookie(c, token)
  c.JSON(http.StatusOK, dto.LoginResponse{User: *user})
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
  token, err := c.Cookie("auth_token")
  if err != nil {
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No hay sesión activa",
    })
    return
  }

  if err := h.service.Logout(token); err != nil {
    middleware.ClearAuthCookie(c)
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No hay sesión activa",
    })
    return
  }

  middleware.ClearAuthCookie(c)
  c.JSON(http.StatusOK, dto.LogoutResponse{Message: "Sesión cerrada exitosamente"})
}

// Me godoc
// @Summary Obtener usuario autenticado
// @Description Retorna los datos del usuario actualmente autenticado, incluyendo su perfil, estado y módulos con permisos (items) asignados
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.MeResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
  userID, exists := middleware.GetUserIDFromContext(c)
  if !exists {
    c.JSON(http.StatusUnauthorized, gin.H{
      "status":  "error",
      "message": "No autenticado",
    })
    return
  }

  user, err := h.service.GetMe(userID)
  if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al cargar datos del usuario",
    })
    return
  }

  c.JSON(http.StatusOK, dto.MeResponse{User: *user})
}

// Refresh godoc
// @Summary Renovar sesión
// @Description Extiende el TTL de la sesión actual
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.RefreshResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
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

  if err := authpkg.ExtendSession(token); err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
      "status":  "error",
      "message": "Error al renovar sesión",
    })
    return
  }

  c.JSON(http.StatusOK, dto.RefreshResponse{Message: "Sesión renovada exitosamente"})
}