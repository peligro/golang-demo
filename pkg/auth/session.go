package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var rdb *redis.Client

// SessionData estructura guardada en Redis
type SessionData struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// InitRedis inicializa la conexión a Redis
func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://redis:6379/0"
	}

	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("error parsing Redis URL: %w", err)
	}

	rdb = redis.NewClient(opts)

	// Test connection
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("error connecting to Redis: %w", err)
	}

	fmt.Println("✅ Conexión exitosa a Redis")
	return nil
}

// GenerateToken genera un token aleatorio seguro (64 bytes = 128 caracteres hex)
func GenerateToken() (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateSession crea una sesión en Redis
func CreateSession(token string, userID uint, email string) error {
	sessionData := SessionData{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("error serializando sesión: %w", err)
	}

	// 🔁 Usar el mismo TTL que la cookie (desde .env)
	ttl := getSessionTTLFromEnv()

	key := fmt.Sprintf("session:%s", token)
	if err := rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("error guardando sesión en Redis: %w", err)
	}

	return nil
}

// GetSession obtiene una sesión de Redis
func GetSession(token string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", token)

	data, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("sesión no encontrada o expirada")
	}
	if err != nil {
		return nil, fmt.Errorf("error obteniendo sesión: %w", err)
	}

	var session SessionData
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("error deserializando sesión: %w", err)
	}

	return &session, nil
}

// DeleteSession elimina una sesión de Redis (logout)
func DeleteSession(token string) error {
	key := fmt.Sprintf("session:%s", token)
	if err := rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("error eliminando sesión: %w", err)
	}
	return nil
}

// ExtendSession extiende el TTL de una sesión
func ExtendSession(token string) error {
	key := fmt.Sprintf("session:%s", token)
	ttl := getSessionTTLFromEnv() // ← Mismo TTL que cookie

	if err := rdb.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("error extendiendo sesión: %w", err)
	}
	return nil
}

// getSessionTTLFromEnv helper para leer TTL desde .env (en segundos)
func getSessionTTLFromEnv() time.Duration {
	ttlStr := os.Getenv("SESSION_TTL")
	if ttlStr == "" {
		return 24 * time.Hour // default
	}
	ttl, err := strconv.Atoi(ttlStr)
	if err != nil || ttl < 300 || ttl > 604800 {
		return 24 * time.Hour
	}
	return time.Duration(ttl) * time.Second
}

// GetRedisClient devuelve el cliente Redis (para uso interno si es necesario)
func GetRedisClient() *redis.Client {
	return rdb
}