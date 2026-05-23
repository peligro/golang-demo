package profile

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

// RegisterRoutes registra todas las rutas de Profile + ProfileModule + ProfileModuleItem
func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)
  moduleHandler := NewModuleHandler(db)
  itemHandler := NewModuleItemHandler(db)

  profiles := router.Group("/profiles")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    profiles.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización base (valida permisos del módulo profiles)
    profiles.Use(middleware.RequireModule(db, common.ModuleProfiles))

    // 👤 CRUD básico de Profiles
    // ✅ Lectura: solo requiere el módulo
    profiles.GET("", handler.Index)
    profiles.GET("/:id", handler.Show)
    
    // 🔐 Escritura/Eliminación: requiere items específicos
    profiles.POST("", middleware.RequireItem(db, common.ModuleProfiles, "crear_perfil"), handler.Create)
    profiles.PUT("/:id", middleware.RequireItem(db, common.ModuleProfiles, "editar_perfil"), handler.Update)
    profiles.DELETE("/:id", middleware.RequireItem(db, common.ModuleProfiles, "eliminar_perfil"), handler.Delete)

    // 🔗 Profile Modules (gestión de módulos asignados a un perfil)
    // ✅ Lectura: solo requiere el módulo base
    profiles.GET("/:id/modules", moduleHandler.Index)
    
    // 🔐 Escritura: requiere item específico para asignar módulos
    profiles.PUT("/:id/modules", middleware.RequireItem(db, common.ModuleProfiles, "asignar_modulos"), moduleHandler.Sync)

    // 🔗 Profile Module Items (gestión de permisos granulares)
    items := profiles.Group("/:id/modules/:moduleId/items")
    {
      // ✅ Lectura: solo requiere el módulo base
      items.GET("", itemHandler.Index)
      
      // 🔐 Escritura: requiere items específicos para gestionar permisos
      items.PUT("", middleware.RequireItem(db, common.ModuleProfiles, "asignar_permisos"), itemHandler.Sync)
      items.POST("", middleware.RequireItem(db, common.ModuleProfiles, "asignar_permisos"), itemHandler.Attach)
      items.DELETE("/:itemId", middleware.RequireItem(db, common.ModuleProfiles, "desasignar_permisos"), itemHandler.Detach)
    }
  }
}