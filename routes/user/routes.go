package user

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

// RegisterRoutes registra las rutas de User en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  users := router.Group("/users")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    users.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización (valida permisos del módulo)
    // Requiere: módulo common.ModuleUsers O item common.ViewAllAdminCode
    users.Use(middleware.RequireModule(db, common.ModuleUsers))

    // ✅ Lectura: solo requiere el módulo
    users.GET("", handler.Index)
    users.GET("/:id", handler.Show)

    // 🔐 Escritura/Eliminación: requiere items específicos (más restrictivo)
    users.POST("", middleware.RequireItem(db, common.ModuleUsers, "crear_usuario"), handler.Create)
    users.PUT("/:id", middleware.RequireItem(db, common.ModuleUsers, "editar_usuario"), handler.Update)
    users.DELETE("/:id", middleware.RequireItem(db, common.ModuleUsers, "eliminar_usuario"), handler.Delete)
    
    // Endpoint especial para usuario autenticado (ya protegido por AuthMiddleware)
    users.GET("/me", handler.Me)
  }
}