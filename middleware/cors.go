
package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware configura headers CORS para permitir peticiones desde el frontend
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		allowedOrigin := getAllowedOrigin(c)
		
		// Permitir solo origen específico (nunca "*" con credenciales)
		c.Header("Access-Control-Allow-Origin", allowedOrigin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400") // 24 horas para cachear preflight

		// Manejar preflight request
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getAllowedOrigin devuelve el origen permitido según el environment
func getAllowedOrigin(c *gin.Context) string {
	env := os.Getenv("ENVIRONMENT")
	
	switch env {
	case "production":
		return getEnv("FRONTEND_URL", "https://tudominio.com")
	case "staging":
		return getEnv("FRONTEND_URL", "https://staging.tudominio.com")
	default:
		// En local, leer el Origin de la request y validar
		origin := c.Request.Header.Get("Origin")
		if isValidLocalOrigin(origin) {
			return origin
		}
		return getEnv("FRONTEND_URL", "http://localhost:5173")
	}
}

// isValidLocalOrigin valida orígenes permitidos en desarrollo
func isValidLocalOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	
	allowed := []string{
		"http://localhost:5173",  // Vite
		"http://localhost:3000",  // Create React App
		"http://127.0.0.1:5173",
		"http://127.0.0.1:3000",
		"http://localhost:5174",
	}
	
	for _, o := range allowed {
		if origin == o {
			return true
		}
	}
	return false
}

// getEnv helper para leer variables de entorno
func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return strings.TrimSpace(val)
	}
	return defaultValue
}