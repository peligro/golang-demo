package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	os.Setenv("ENVIRONMENT", "local") // Asegura que HSTS no se active en este test

	router := gin.New()
	router.Use(SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// ✅ Headers obligatorios
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.Equal(t, "no-store, no-cache, must-revalidate, private", w.Header().Get("Cache-Control"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'none'")
	assert.Contains(t, w.Header().Get("Permissions-Policy"), "camera=()")

	// ✅ Headers sensibles eliminados
	assert.Empty(t, w.Header().Get("Server"))
	assert.Empty(t, w.Header().Get("X-Powered-By"))
	assert.Empty(t, w.Header().Get("X-AspNet-Version"))
}
