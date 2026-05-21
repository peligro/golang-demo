package auth

import (
  "github.com/gin-gonic/gin"
  "github.com/peligro/golang-demo/middleware"
  "gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
  handler := NewHandler(db)

  auth := router.Group("/auth")
  {
    // Rutas públicas
    auth.POST("/login", handler.Login)
    auth.POST("/logout", handler.Logout)

    // Rutas protegidas (con validación de state)
    auth.GET("/me", middleware.AuthMiddleware(db), handler.Me)
    auth.POST("/refresh", middleware.AuthMiddleware(db), handler.Refresh)
  }
}