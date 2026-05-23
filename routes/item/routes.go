package item

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

// RegisterRoutes registra las rutas de Item en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  items := router.Group("/items")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    items.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización (valida permisos del módulo)
    // Requiere: módulo common.ModuleItems O item common.ViewAllAdminCode
    items.Use(middleware.RequireModule(db, common.ModuleItems))

    // ✅ Lectura: solo requiere el módulo
    items.GET("", handler.Index)
    items.GET("/:id", handler.Show)

    // 🔐 Escritura/Eliminación: requiere items específicos (más restrictivo)
    items.POST("", middleware.RequireItem(db, common.ModuleItems, "crear_item"), handler.Create)
    items.PUT("/:id", middleware.RequireItem(db, common.ModuleItems, "editar_item"), handler.Update)
    items.DELETE("/:id", middleware.RequireItem(db, common.ModuleItems, "eliminar_item"), handler.Delete)
  }
}