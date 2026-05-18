package state

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registra las rutas de State en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	handler := NewHandler(db)
	
	states := router.Group("/states")
	{
		states.GET("", handler.Index)
		states.GET("/:id", handler.Show)
		states.POST("", handler.Create)
		states.PUT("/:id", handler.Update)
		states.DELETE("/:id", handler.Delete)
	}
}
