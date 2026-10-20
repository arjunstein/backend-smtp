package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	payload := map[string]string{"foo": "bar"}

	Success(w, http.StatusOK, "Operation successful", payload)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var res APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Operation successful", res.Message)
	assert.Nil(t, res.Error)

	dataMap, ok := res.Data.(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "bar", dataMap["foo"])
}

func TestBadRequestResponse(t *testing.T) {
	w := httptest.NewRecorder()
	errSample := errors.New("invalid parameter")

	BadRequest(w, "Bad input", errSample)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var res APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Bad input", res.Message)
	assert.Equal(t, "invalid parameter", res.Error)
	assert.Nil(t, res.Data)
}

func TestNotFoundResponse(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "Resource not found")

	assert.Equal(t, http.StatusNotFound, w.Code)

	var res APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Resource not found", res.Message)
}

func TestUnauthorizedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w, "Unauthorized access")

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var res APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Unauthorized access", res.Message)
}

func TestInternalServerErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	InternalServerError(w, "Database failure", "connection timeout")

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var res APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Database failure", res.Message)
	assert.Equal(t, "connection timeout", res.Error)
}
