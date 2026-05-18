package app

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/peligro/golang-demo/database"
	"github.com/peligro/golang-demo/middleware"
	"github.com/peligro/golang-demo/model"
	
	// ← Imports actualizados con subpaquetes
	"github.com/peligro/golang-demo/routes/health"
	"github.com/peligro/golang-demo/routes/state"

	_ "github.com/peligro/golang-demo/docs"
)

func SetupRouter() *gin.Engine {
	// ... (carga de env, modo Gin, DB, migraciones, middlewares) ...

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	// Rutas públicas
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"mensaje": "Hola mundo desde Golang con Gin Framework",
		})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"estado":  "error",
			"message": "Recurso no disponible",
			"code":    http.StatusNotFound,
		})
	})

	// 🩺 Health check (sin DB)
	health.RegisterRoutes(router)

	// 🗂️ States CRUD (con DB)
	state.RegisterRoutes(router, db)

	// Swagger docs
	if environment == "local" || environment == "staging" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}