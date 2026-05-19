package item

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registra las rutas de Item en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	handler := NewHandler(db)

	items := router.Group("/items")
	{
		items.GET("", handler.Index)
		items.GET("/:id", handler.Show)
		items.POST("", handler.Create)
		items.PUT("/:id", handler.Update)
		items.DELETE("/:id", handler.Delete)
	}
}