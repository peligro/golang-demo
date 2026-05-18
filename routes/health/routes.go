package health

import "github.com/gin-gonic/gin"

// RegisterRoutes registra las rutas de Health en el router
func RegisterRoutes(router *gin.Engine) {
	router.GET("/health", Index)
}
