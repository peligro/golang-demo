package app_menu

import (
  "github.com/gin-gonic/gin"
  "gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  // Ruta especial sin paginación (para sidebar)
  router.GET("/app-menu-all", handler.ListAll)

  // Grupo CRUD estándar
  group := router.Group("/app-menu")
  {
    group.GET("", handler.List)
    group.GET("/:id", handler.Get)
    group.POST("", handler.Create)
    group.PUT("/:id", handler.Update)
    group.DELETE("/:id", handler.Delete)
  }
}
