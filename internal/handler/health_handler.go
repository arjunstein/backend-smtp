package handler

import (
	"net/http"
	"time"

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
// @Success 200 {object} response.APIResponse "Service is healthy and ready to dispatch emails"
// @Router /api/v1/health [get]
// @Router /health [get]
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, "Service is healthy", map[string]any{
		"status":    "UP",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "backend-smtp",
	})
}
