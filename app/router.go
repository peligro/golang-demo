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
	"github.com/peligro/golang-demo/routes/module"
	"github.com/peligro/golang-demo/routes/profile"
	"github.com/peligro/golang-demo/routes/item"
	"github.com/peligro/golang-demo/routes/user"
	"github.com/peligro/golang-demo/routes/auth"

	_ "github.com/peligro/golang-demo/docs"
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

	// 🧩 Modules CRUD (con DB)
	module.RegisterRoutes(router, db)

	// 👥 Profiles CRUD (con DB)
	profile.RegisterRoutes(router, db)

	// 🧩 Items CRUD (con DB)
	item.RegisterRoutes(router, db)

	// 👤 Users CRUD (con DB + bcrypt)
	user.RegisterRoutes(router, db)

	// 🔐 Auth routes
	auth.RegisterRoutes(router, db)

	if environment == "local" || environment == "staging" {
    // Redirigir /swagger a /swagger/index.html?dark=true
    router.GET("/swagger", func(c *gin.Context) {
        c.Redirect(http.StatusMovedPermanently, "/swagger/index.html?dark=true")
    })
    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	return router
}