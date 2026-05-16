package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	
)
// Health_check godoc
// @Summary Verifica estado del servidor
// @Description Retorna "UP!!" si el servicio está activo
// @Tags Health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /health [get]
func Health_index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "UP!!"})
}