package profile

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registra todas las rutas de Profile + ProfileModule + ProfileModuleItem
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	// Handlers
	handler := NewHandler(db)
	moduleHandler := NewModuleHandler(db)
	itemHandler := NewModuleItemHandler(db)

	profiles := router.Group("/profiles")
	{
		// 👤 CRUD básico de Profiles
		profiles.GET("", handler.Index)
		profiles.GET("/:id", handler.Show)
		profiles.POST("", handler.Create)
		profiles.PUT("/:id", handler.Update)
		profiles.DELETE("/:id", handler.Delete)

		// 🔗 Profile Modules (anidado)
		profiles.GET("/:profileId/modules", moduleHandler.Index)
		profiles.PUT("/:profileId/modules", moduleHandler.Sync)

		// 🔗 Profile Module Items (anidado profundo)
		items := profiles.Group("/:profileId/modules/:moduleId/items")
		{
			items.GET("", itemHandler.Index)
			items.PUT("", itemHandler.Sync)
			items.POST("", itemHandler.Attach)
			items.DELETE("/:itemId", itemHandler.Detach)
		}
	}
}