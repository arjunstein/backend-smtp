package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/arjunstein/backend-smtp/pkg/response"
)

type EmailHandler struct {
	emailService service.EmailService
}

func NewEmailHandler(emailService service.EmailService) *EmailHandler {
	return &EmailHandler{emailService: emailService}
}

// Send godoc
// @Summary Send email via centralized SMTP agent
// @Description Dispatches an email using the specified SMTP config or the system default
// @Tags Emails
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param payload body model.SendEmailRequest true "Send email payload"
// @Success 200 {object} response.APIResponse{data=model.SendEmailResponse} "Email successfully dispatched"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or mail delivery failure"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/emails/send [post]
func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	var req model.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	res, err := h.emailService.SendEmail(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrNoActiveSMTPConfig) || errors.Is(err, service.ErrSMTPConfigInactive) {
			response.BadRequest(w, err.Error(), nil)
			return
		}
		response.BadRequest(w, "Failed to send email: "+err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "Email sent successfully", res)
}
