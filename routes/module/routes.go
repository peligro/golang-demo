package module

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/common"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  modules := router.Group("/modules")
  {
    // 🔐 Capa 1: Autenticación (valida cookie + estado activo)
    modules.Use(middleware.AuthMiddleware(db))
    
    // 🔐 Capa 2: Autorización RESTRICTIVA (módulos es crítico)
    // Requiere: item "gestionar_modulos" O item "view_all_admin"
    // NOTA: No basta con tener el módulo, se necesita permiso explícito
    modules.Use(middleware.RequireItem(db, common.ModuleModules, "gestionar_modulos"))

    // ✅ Todas las operaciones requieren el mismo nivel de permiso
    // (lectura también es sensible porque revela la estructura de permisos)
    modules.GET("", handler.Index)
    modules.GET("/:id", handler.Show)
    modules.POST("", handler.Create)
    modules.PUT("/:id", handler.Update)
    modules.DELETE("/:id", handler.Delete)
  }
}