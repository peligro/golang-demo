package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware aplica headers de seguridad OWASP
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		env := os.Getenv("ENVIRONMENT")
		isProd := env == "production" || env == "staging"

		// 🔒 Anti-cache para APIs
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// 🔒 OWASP Security Headers
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Permissions-Policy
		c.Header("Permissions-Policy",
			"accelerometer=(), camera=(), geolocation=(), gyroscope=(), "+
				"magnetometer=(), microphone=(), payment=(), usb=()")

		// COOP/COEP - Protección Spectre/Meltdown
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// CSP - Ajustado para API + Swagger
		csp := "default-src 'none'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self';"
		c.Header("Content-Security-Policy", csp)

		// HSTS - Solo producción/staging con HTTPS
		if isProd {
			c.Header("Strict-Transport-Security", 
				"max-age=31536000; includeSubDomains; preload")
		}

		c.Next()

		// 🔒 Eliminar headers sensibles después de procesar
		h := c.Writer.Header()
		h.Del("Server")
		h.Del("X-Powered-By")
		h.Del("X-AspNet-Version")
		h.Del("X-AspNetMvc-Version")
	}
}

// Helper para CSP dinámico
func buildCSP(allowedOrigins []string) string {
	connectSrc := "connect-src 'self'"
	if len(allowedOrigins) > 0 {
		connectSrc += " " + strings.Join(allowedOrigins, " ")
	}
	
	return "default-src 'none'; " +
		"script-src 'self' 'unsafe-inline'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data:; " +
		"font-src 'self'; " +
		connectSrc + "; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self';"
}