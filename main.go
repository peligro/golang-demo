package main

import (
	"github.com/peligro/golang-demo/app"
	"github.com/peligro/golang-demo/pkg/auth"
	"github.com/peligro/golang-demo/common"
	"log"
	"os"

	
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var version string = "dev"

// @title Golang DEMO
// @version 1.0
// @description API Backend para GOLANG DEMO.
// @host localhost:8082
// @BasePath /
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// Para generar la documentación:
// docker exec -it go-dev swag init --parseDependency --parseInternal
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 🔐 Inicializar conexión a Redis para sesiones
	if err := auth.InitRedis(); err != nil {
		log.Fatalf("❌ Error al inicializar Redis: %v", err)
	}
	log.Println("✅ Redis conectado para gestión de sesiones")

	// 🔥 Registrar validaciones personalizadas ANTES de usar el router
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		common.SetupValidator(v)
		log.Println("✅ Validaciones personalizadas registradas")
	}

	// Configurar router con todas las rutas
	router := app.SetupRouter()

	log.Printf("🚀 Iniciando servidor %s en puerto %s", version, port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}