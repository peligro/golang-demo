package profile

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registra las rutas de Profile en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	handler := NewHandler(db)

	profiles := router.Group("/profiles")
	{
		profiles.GET("", handler.Index)
		profiles.GET("/:id", handler.Show)
		profiles.POST("", handler.Create)
		profiles.PUT("/:id", handler.Update)
		profiles.DELETE("/:id", handler.Delete)
	}
}