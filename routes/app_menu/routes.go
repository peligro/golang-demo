package app_menu

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  // 🌐 Ruta PÚBLICA: para construir el sidebar dinámico en el frontend
  // No requiere autenticación ni permisos
  router.GET("/app-menu-all", handler.ListAll)

  // 🔐 Grupo CRUD protegido
  group := router.Group("/app-menu")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    group.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización (valida permisos del módulo)
    // Requiere: módulo common.ModuleAppMenu O item common.ViewAllAdminCode
    group.Use(middleware.RequireModule(db, common.ModuleAppMenu))

    // ✅ Lectura: solo requiere el módulo
    group.GET("", handler.List)
    group.GET("/:id", handler.Get)

    // 🔐 Escritura/Eliminación: requiere items específicos (más restrictivo)
    group.POST("", middleware.RequireItem(db, common.ModuleAppMenu, "crear_menu"), handler.Create)
    group.PUT("/:id", middleware.RequireItem(db, common.ModuleAppMenu, "editar_menu"), handler.Update)
    group.DELETE("/:id", middleware.RequireItem(db, common.ModuleAppMenu, "eliminar_menu"), handler.Delete)
  }
}