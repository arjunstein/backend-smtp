package response

import (
	"encoding/json"
	"net/http"
)

// APIResponse represents standard uniform HTTP response envelope.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// JSON writes standard JSON response to http.ResponseWriter.
func JSON(w http.ResponseWriter, statusCode int, success bool, message string, data any, err any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var errVal any = err
	if e, ok := err.(error); ok && e != nil {
		errVal = e.Error()
	}

	_ = json.NewEncoder(w).Encode(APIResponse{
		Success: success,
		Message: message,
		Data:    data,
		Error:   errVal,
	})
}

// Success writes successful response with custom status code and payload.
func Success(w http.ResponseWriter, statusCode int, message string, data any) {
	JSON(w, statusCode, true, message, data, nil)
}

// Error writes failure response with custom status code and error details.
func Error(w http.ResponseWriter, statusCode int, message string, err any) {
	JSON(w, statusCode, false, message, nil, err)
}

// BadRequest writes 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string, err any) {
	Error(w, http.StatusBadRequest, message, err)
}

// Unauthorized writes 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, message, nil)
}

// NotFound writes 404 Not Found response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, message, nil)
}

// InternalServerError writes 500 Internal Server Error response.
func InternalServerError(w http.ResponseWriter, message string, err any) {
	Error(w, http.StatusInternalServerError, message, err)
}
