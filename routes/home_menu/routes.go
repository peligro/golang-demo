package home_menu

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  // 🌐 Ruta PÚBLICA: para construir el dashboard/home sin autenticación
  // No requiere auth ni permisos (puede ser caché público)
  router.GET("/home-menu-all", handler.ListAll)

  // 🔐 Grupo CRUD protegido (panel de administración)
  group := router.Group("/home-menu")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    group.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización (valida permisos)
    // Requiere: módulo common.ModuleHome O item common.ViewAllAdminCode
    // Nota: home_menu es más flexible, puede no requerir módulo específico si quieres
    group.Use(middleware.RequireModule(db, common.ModuleHome))

    // ✅ Lectura: solo requiere el módulo
    group.GET("", handler.List)
    group.GET("/:id", handler.Get)

    // 🔐 Escritura/Eliminación: requiere items específicos (más restrictivo)
    group.POST("", middleware.RequireItem(db, common.ModuleHome, "crear_tarjeta"), handler.Create)
    group.PUT("/:id", middleware.RequireItem(db, common.ModuleHome, "editar_tarjeta"), handler.Update)
    group.DELETE("/:id", middleware.RequireItem(db, common.ModuleHome, "eliminar_tarjeta"), handler.Delete)
  }
}