package middleware

import (
  "net/http"
  "os"
  "strconv"

  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/model"
  "github.com/peligro/golang-demo/pkg/auth"
  "gorm.io/gorm"
)

// =============================================================================
// COOKIE HELPERS
// =============================================================================

func getCookieDomain() string {
  env := os.Getenv("ENVIRONMENT")
  if env == "production" || env == "staging" {
    domain := os.Getenv("COOKIE_DOMAIN")
    if domain != "" {
      return domain
    }
  }
  return ""
}

func getSessionTTL() int {
  ttlStr := os.Getenv("SESSION_TTL")
  if ttlStr == "" {
    return 86400
  }
  ttl, err := strconv.Atoi(ttlStr)
  if err != nil || ttl < 300 || ttl > 604800 {
    return 86400
  }
  return ttl
}

func SetAuthCookie(c *gin.Context, token string) {
  secure := false
  env := os.Getenv("ENVIRONMENT")
  if env == "production" || env == "staging" {
    secure = true
  }
  if os.Getenv("COOKIE_SECURE") == "true" {
    secure = true
  }

  c.SetCookie(
    "auth_token",
    token,
    getSessionTTL(),
    "/",
    getCookieDomain(),
    secure,
    true,
  )
}

func ClearAuthCookie(c *gin.Context) {
  c.SetCookie(
    "auth_token",
    "",
    -1,
    "/",
    getCookieDomain(),
    false,
    true,
  )
}

// =============================================================================
// AUTH MIDDLEWARES
// =============================================================================

// AuthMiddleware verifica la cookie de autenticación y que el usuario esté activo
func AuthMiddleware(db *gorm.DB) gin.HandlerFunc {
  return func(c *gin.Context) {
    // Obtener token de la cookie
    token, err := c.Cookie("auth_token")
    if err != nil {
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "No autenticado",
      })
      c.Abort()
      return
    }

    // Validar sesión en Redis
    session, err := auth.GetSession(token)
    if err != nil {
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "Sesión inválida o expirada",
      })
      c.Abort()
      return
    }

    // 🔍 Verificar que el usuario exista y esté activo en DB
    var meta model.UserMetadata
    if err := db.Where("user_id = ?", session.UserID).First(&meta).Error; err != nil {
      // Usuario no encontrado → invalidar sesión
      _ = auth.DeleteSession(token)
      ClearAuthCookie(c)
      
      c.JSON(http.StatusUnauthorized, gin.H{
        "status":  "error",
        "message": "Usuario no encontrado",
      })
      c.Abort()
      return
    }

    // 🔒 Validar estado activo
    if meta.State != 1 {
      // Usuario inactivo → invalidar sesión
      _ = auth.DeleteSession(token)
      ClearAuthCookie(c)
      
      c.JSON(http.StatusForbidden, gin.H{
        "status":  "error",
        "message": "Usuario inactivo. Contacta al administrador",
        "state":   meta.State,
      })
      c.Abort()
      return
    }

    // ✅ Todo OK: inyectar datos en contexto
    c.Set("user_id", session.UserID)
    c.Set("user_email", session.Email)
    c.Set("session", session)
    c.Set("user_state", meta.State)
    c.Set("user_profile_id", meta.ProfileID)

    c.Next()
  }
}

// OptionalAuthMiddleware permite acceso anónimo pero inyecta user si existe
func OptionalAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
  return func(c *gin.Context) {
    token, err := c.Cookie("auth_token")
    if err != nil {
      c.Next()
      return
    }

    session, err := auth.GetSession(token)
    if err != nil {
      c.Next()
      return
    }

    // Verificar estado (sin bloquear)
    var meta model.UserMetadata
    if err := db.Where("user_id = ?", session.UserID).First(&meta).Error; err == nil && meta.State == 1 {
      c.Set("user_id", session.UserID)
      c.Set("user_email", session.Email)
      c.Set("session", session)
    }

    c.Next()
  }
}

// =============================================================================
// CONTEXT HELPERS
// =============================================================================

func GetUserIDFromContext(c *gin.Context) (uint, bool) {
  userID, exists := c.Get("user_id")
  if !exists {
    return 0, false
  }
  return userID.(uint), true
}

func GetUserEmailFromContext(c *gin.Context) (string, bool) {
  email, exists := c.Get("user_email")
  if !exists {
    return "", false
  }
  return email.(string), true
}

func GetUserStateFromContext(c *gin.Context) (int, bool) {
  state, exists := c.Get("user_state")
  if !exists {
    return 0, false
  }
  return state.(int), true
}