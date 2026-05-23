package state

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

// RegisterRoutes registra las rutas de State en el router
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  states := router.Group("/states")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    states.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización (valida permisos del módulo states)
    // Requiere: módulo common.ModuleStates O item common.ViewAllAdminCode
    states.Use(middleware.RequireModule(db, common.ModuleStates))

    // ✅ Lectura: solo requiere el módulo
    states.GET("", handler.Index)
    states.GET("/:id", handler.Show)

    // 🔐 Escritura/Eliminación: requiere items específicos (más restrictivo)
    states.POST("", middleware.RequireItem(db, common.ModuleStates, "crear_estado"), handler.Create)
    states.PUT("/:id", middleware.RequireItem(db, common.ModuleStates, "editar_estado"), handler.Update)
    states.DELETE("/:id", middleware.RequireItem(db, common.ModuleStates, "eliminar_estado"), handler.Delete)
  }
}