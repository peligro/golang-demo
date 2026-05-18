package health

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/peligro/golang-demo/dto"
)

// Index godoc
// @Summary Verifica estado del servidor
// @Description Retorna "UP!!" si el servicio está activo
// @Tags Health
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Router /health [get]
func Index(c *gin.Context) {  // ← Renombrado de Health_index a Index
	c.JSON(http.StatusOK, dto.HealthResponse{Status: "UP!!"})
}