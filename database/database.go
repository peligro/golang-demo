package database

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// GetDB devuelve una instancia de *gorm.DB configurada y lista para usar.
func GetDB() *gorm.DB {
	// Cargar .env si existe (solo para desarrollo local)
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠️  No se encontró .env, usando variables de entorno del sistema")
	}

	// Obtener entorno
	environment := os.Getenv("ENVIRONMENT")
	isLocal := environment == "local"

	// Validar variables críticas
	host := os.Getenv("DB_HOST")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	database := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	timezone := os.Getenv("TIMEZONE")

	if host == "" || username == "" || database == "" {
		panic("❌ Variables de entorno de base de datos incompletas")
	}
	if port == "" {
		port = "5432"
	}
	if timezone == "" {
		timezone = "UTC"
	}

	// Configurar sslmode según entorno
	sslMode := "require" // Producción
	if isLocal {
		sslMode = "disable"
	}

	// Construir DSN
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		host, username, password, database, port, sslMode, timezone,
	)

	// Conectar con GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("❌ Error al conectar con PostgreSQL: %v", err))
	}

	// Configurar Pool de Conexiones nativo
	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("❌ Error obteniendo sql.DB: %v", err))
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	status := "✅ Conexión exitosa a PostgreSQL"
	if !isLocal {
		status += " (con SSL)"
	}
	fmt.Println(status)

	return db
}