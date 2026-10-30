package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/arjunstein/backend-smtp/internal/model"
	"github.com/arjunstein/backend-smtp/internal/repository"
	"github.com/arjunstein/backend-smtp/internal/service"
	"github.com/arjunstein/backend-smtp/pkg/response"
	"github.com/go-chi/chi/v5"
)

type SMTPConfigHandler struct {
	service service.SMTPConfigService
}

func NewSMTPConfigHandler(service service.SMTPConfigService) *SMTPConfigHandler {
	return &SMTPConfigHandler{service: service}
}

// Create godoc
// @Summary Create new SMTP configuration
// @Description Registers a new SMTP account for central email delivery
// @Tags SMTP Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param payload body model.CreateSMTPConfigRequest true "SMTP configuration payload"
// @Success 201 {object} response.APIResponse{data=model.SMTPConfigResponse} "SMTP configuration created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload or validation error"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 409 {object} response.ConflictResponse "Configuration name already exists"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs [post]
func (h *SMTPConfigHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSMTPConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	res, err := h.service.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrDuplicateConfigName) {
			response.Error(w, http.StatusConflict, err.Error(), nil)
			return
		}
		response.BadRequest(w, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, "SMTP configuration created successfully", res)
}

// List godoc
// @Summary List all SMTP configurations
// @Description Returns all registered SMTP configurations with masked credentials
// @Tags SMTP Config
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Success 200 {object} response.APIResponse{data=[]model.SMTPConfigResponse} "List of SMTP configurations"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs [get]
func (h *SMTPConfigHandler) List(w http.ResponseWriter, r *http.Request) {
	configs, err := h.service.FindAll(r.Context())
	if err != nil {
		response.InternalServerError(w, "Failed to retrieve configurations", err)
		return
	}

	response.Success(w, http.StatusOK, "SMTP configurations retrieved successfully", configs)
}

// GetByID godoc
// @Summary Get SMTP configuration by ID
// @Description Fetches specific SMTP configuration details by numeric ID
// @Tags SMTP Config
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path int true "SMTP Config ID"
// @Success 200 {object} response.APIResponse{data=model.SMTPConfigResponse} "SMTP configuration details"
// @Failure 400 {object} response.ErrorResponse "Invalid ID format"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 404 {object} response.NotFoundResponse "Configuration not found"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs/{id} [get]
func (h *SMTPConfigHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.BadRequest(w, "Invalid config ID in URL parameter", err)
		return
	}

	cfg, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrSMTPConfigNotFound) {
			response.NotFound(w, "SMTP configuration not found")
			return
		}
		response.InternalServerError(w, "Failed to retrieve configuration", err)
		return
	}

	response.Success(w, http.StatusOK, "SMTP configuration retrieved successfully", cfg)
}

// Update godoc
// @Summary Update SMTP configuration
// @Description Updates fields of an existing SMTP configuration
// @Tags SMTP Config
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path int true "SMTP Config ID"
// @Param payload body model.UpdateSMTPConfigRequest true "Update payload"
// @Success 200 {object} response.APIResponse{data=model.SMTPConfigResponse} "SMTP configuration updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid input data or validation error"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 404 {object} response.NotFoundResponse "Configuration not found"
// @Failure 409 {object} response.ConflictResponse "Configuration name conflict"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs/{id} [put]
func (h *SMTPConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.BadRequest(w, "Invalid config ID in URL parameter", err)
		return
	}

	var req model.UpdateSMTPConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid JSON payload", err)
		return
	}

	res, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrSMTPConfigNotFound) {
			response.NotFound(w, "SMTP configuration not found")
			return
		}
		if errors.Is(err, service.ErrDuplicateConfigName) {
			response.Error(w, http.StatusConflict, err.Error(), nil)
			return
		}
		response.BadRequest(w, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "SMTP configuration updated successfully", res)
}

// Delete godoc
// @Summary Delete SMTP configuration
// @Description Deletes an SMTP configuration by ID
// @Tags SMTP Config
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path int true "SMTP Config ID"
// @Success 200 {object} response.SimpleResponse "SMTP configuration deleted"
// @Failure 400 {object} response.ErrorResponse "Invalid ID format"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 404 {object} response.NotFoundResponse "Configuration not found"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs/{id} [delete]
func (h *SMTPConfigHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.BadRequest(w, "Invalid config ID in URL parameter", err)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrSMTPConfigNotFound) {
			response.NotFound(w, "SMTP configuration not found")
			return
		}
		response.InternalServerError(w, "Failed to delete configuration", err)
		return
	}

	response.Success(w, http.StatusOK, "SMTP configuration deleted successfully", nil)
}

// SetDefault godoc
// @Summary Set SMTP configuration as system default
// @Description Sets target configuration as active default for outbound email fallback
// @Tags SMTP Config
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path int true "SMTP Config ID"
// @Success 200 {object} response.SimpleResponse "Configuration set as default"
// @Failure 400 {object} response.ErrorResponse "Invalid ID"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 404 {object} response.NotFoundResponse "Configuration not found"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs/{id}/default [put]
func (h *SMTPConfigHandler) SetDefault(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.BadRequest(w, "Invalid config ID in URL parameter", err)
		return
	}

	if err := h.service.SetDefault(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrSMTPConfigNotFound) {
			response.NotFound(w, "SMTP configuration not found")
			return
		}
		response.InternalServerError(w, "Failed to set default configuration", err)
		return
	}

	response.Success(w, http.StatusOK, "SMTP configuration marked as default successfully", nil)
}

// TestConnection godoc
// @Summary Test SMTP connection
// @Description Pings and tests authentication with the specified SMTP host
// @Tags SMTP Config
// @Produce json
// @Security ApiKeyAuth
// @Security BearerAuth
// @Param id path int true "SMTP Config ID"
// @Success 200 {object} response.APIResponse{data=model.TestConnectionResponse} "SMTP connection successful"
// @Failure 400 {object} response.ErrorResponse "Invalid ID or SMTP handshake failure"
// @Failure 401 {object} response.UnauthorizedResponse "Invalid or missing API authentication credentials"
// @Failure 404 {object} response.NotFoundResponse "Configuration not found"
// @Failure 500 {object} response.InternalErrorResponse "Internal server error"
// @Router /api/v1/smtp-configs/{id}/test [post]
func (h *SMTPConfigHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.BadRequest(w, "Invalid config ID in URL parameter", err)
		return
	}

	if err := h.service.TestConnection(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrSMTPConfigNotFound) {
			response.NotFound(w, "SMTP configuration not found")
			return
		}
		response.BadRequest(w, "SMTP connection test failed: "+err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, "SMTP connection test succeeded", map[string]string{
		"status": "CONNECTED",
	})
}

func parseID(r *http.Request) (uint, error) {
	param := chi.URLParam(r, "id")
	val, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}
