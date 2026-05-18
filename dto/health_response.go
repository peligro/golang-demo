package dto

// HealthResponse respuesta del health check
type HealthResponse struct {
	Status string `json:"status" example:"UP!!"`
}
