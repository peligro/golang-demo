package main

import (
	"github.com/peligro/golang-demo/app"
	"log"
	"os"
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


//para generar la documentación docker exec -it go-dev swag init --parseDependency --parseInternal
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := app.SetupRouter()

	log.Printf("🚀 Iniciando servidor %s en puerto %s", version, port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("❌ Error al iniciar el servidor: %v", err)
	}
}

 