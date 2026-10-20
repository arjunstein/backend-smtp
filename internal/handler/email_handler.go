package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/arjunstein/backend-smtp/pkg/response"
	"gorm.io/gorm"
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
// @Param payload body model.SendEmailRequest true "Send email payload"
// @Success 200 {object} response.APIResponse "Email successfully dispatched"
// @Failure 400 {object} response.APIResponse "Invalid request payload or mail delivery failure"
// @Failure 401 {object} response.APIResponse "Unauthorized"
// @Failure 500 {object} response.APIResponse "Internal server error"
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

// ListLogs godoc
// @Summary List email delivery audit logs
// @Description Returns paginated history of email dispatches
// @Tags Emails
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Items per page (max 100)" default(20)
// @Param offset query int false "Item offset" default(0)
// @Success 200 {object} response.APIResponse "Email audit logs"
// @Failure 401 {object} response.APIResponse "Unauthorized"
// @Failure 500 {object} response.APIResponse "Internal server error"
// @Router /api/v1/emails/logs [get]
func (h *EmailHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	limit := 20
	offset := 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	logs, total, err := h.emailService.GetEmailLogs(r.Context(), limit, offset)
	if err != nil {
		response.InternalServerError(w, "Failed to retrieve email logs", err)
		return
	}

	response.Success(w, http.StatusOK, "Email logs retrieved successfully", map[string]any{
		"items":  logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetLogByID godoc
// @Summary Get email delivery audit log by ID
// @Description Returns audit log details for a single email dispatch
// @Tags Emails
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Email Log ID"
// @Success 200 {object} response.APIResponse "Email log details"
// @Failure 400 {object} response.APIResponse "Invalid ID format"
// @Failure 401 {object} response.APIResponse "Unauthorized"
// @Failure 404 {object} response.APIResponse "Log not found"
// @Failure 500 {object} response.APIResponse "Internal server error"
// @Router /api/v1/emails/logs/{id} [get]
func (h *EmailHandler) GetLogByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.BadRequest(w, "Invalid log ID in URL parameter", err)
		return
	}

	logItem, err := h.emailService.GetEmailLogByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(w, "Email log not found")
			return
		}
		response.InternalServerError(w, "Failed to retrieve email log", err)
		return
	}

	response.Success(w, http.StatusOK, "Email log retrieved successfully", logItem)
}
