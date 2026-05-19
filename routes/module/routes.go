package module

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	handler := NewHandler(db)

	modules := router.Group("/modules")
	{
		modules.GET("", handler.Index)
		modules.GET("/:id", handler.Show)
		modules.POST("", handler.Create)
		modules.PUT("/:id", handler.Update)
		modules.DELETE("/:id", handler.Delete)
	}
}