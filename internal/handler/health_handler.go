package handler

import (
	"net/http"

	"github.com/arjunstein/backend-smtp/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// CheckHealth godoc
// @Summary Service health check
// @Description Returns uptime status and health of the central SMTP sender service
// @Tags Health
// @Produce json
// @Success 200 {object} response.HealthResponse "Service is healthy"
// @Router /health [get]
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, "Service is healthy", nil)
}
