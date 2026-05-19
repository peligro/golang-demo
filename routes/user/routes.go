package user

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registra las rutas de User en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	handler := NewHandler(db)

	users := router.Group("/users")
	{
		users.GET("", handler.Index)              // Listar con paginación
		users.GET("/:id", handler.Show)           // Ver uno
		users.POST("", handler.Create)            // Crear
		users.PUT("/:id", handler.Update)         // Actualizar
		users.DELETE("/:id", handler.Delete)      // Eliminar
		users.GET("/me", handler.Me)              // Usuario autenticado (futuro)
	}
}