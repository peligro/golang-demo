package app

import (
	"github.com/peligro/golang-demo/middleware"
	"github.com/peligro/golang-demo/routes"
	"github.com/peligro/golang-demo/database"
	"github.com/peligro/golang-demo/model"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	
)

func SetupRouter() *gin.Engine {
	// Cargar .env (útil si main.go no lo hizo, aunque ahora sí lo hace)
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No se encontró .env, usando variables de entorno del sistema")
	}

	environment := os.Getenv("ENVIRONMENT")
	if environment == "local" || environment == "staging" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	// 👇 Ahora pasamos la DB a Migraciones
	db := database.GetDB()
	if environment == "local" {
    log.Println("🔄 Ejecutando migraciones automáticas (dev/staging)...")
    if err := model.Migrations(db); err != nil {
        log.Printf("⚠️  Warning en migraciones: %v", err)
        // Opcional: panic si quieres que falle el inicio en dev
        // panic(err)
    }
	} else {
		log.Println("🔒 Producción: migraciones deben ejecutarse vía pipeline")
	}

	// 1️⃣ CORS primero (maneja OPTIONS)
	router.Use(middleware.CORSMiddleware())
	
	// 2️⃣ Security headers después
	router.Use(middleware.SecurityHeadersMiddleware())
	

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"mensaje":  "Hola mundo desde Golang con Gin Framework",
		})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"estado": "error", "message": "Recurso no disponible"})
	})

	// Rutas de la API
	router.GET("/health", routes.Health_index)
	 
 

	return router
}
