package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlosealves2/video-ia/service-discover/internal/bootstrap"
	"github.com/carlosealves2/video-ia/service-discover/internal/config"
	"github.com/carlosealves2/video-ia/service-discover/internal/domain"
)

func setupTestApp() *gin.Engine {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{
		Port:     8080,
		LogLevel: "error",
		GinMode:  "test",
	}

	app := bootstrap.New(cfg).
		InitLogger().
		InitRepository().
		InitHandlers().
		InitRouter()

	return app.GetRouter()
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestRegisterService(t *testing.T) {
	router := setupTestApp()

	service := domain.RegisterServiceRequest{
		Name: "test-service",
		Host: "localhost",
		Port: 3000,
		Tags: []string{"test"},
	}

	body, _ := json.Marshal(service)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response domain.Service
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-service", response.Name)
	assert.Equal(t, "localhost", response.Host)
	assert.Equal(t, 3000, response.Port)
	assert.NotEmpty(t, response.ID)
}

func TestRegisterServiceInvalidRequest(t *testing.T) {
	router := setupTestApp()

	body := []byte(`{"name": "test"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListServices(t *testing.T) {
	router := setupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/list", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "services")
	assert.Contains(t, response, "count")
}

func TestGetServiceNotFound(t *testing.T) {
	router := setupTestApp()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/services/non-existent-id", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestServiceLifecycle(t *testing.T) {
	router := setupTestApp()

	// Register
	service := domain.RegisterServiceRequest{
		Name: "lifecycle-service",
		Host: "127.0.0.1",
		Port: 4000,
	}

	body, _ := json.Marshal(service)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/services/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)

	var registered domain.Service
	err := json.Unmarshal(w.Body.Bytes(), &registered)
	require.NoError(t, err)
	serviceID := registered.ID

	// Get
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/services/"+serviceID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	update := domain.UpdateServiceRequest{
		Port: 5000,
	}
	body, _ = json.Marshal(update)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/services/"+serviceID+"/update", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updated domain.Service
	err = json.Unmarshal(w.Body.Bytes(), &updated)
	require.NoError(t, err)
	assert.Equal(t, 5000, updated.Port)

	// Heartbeat
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/services/"+serviceID+"/heartbeat", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Unregister
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/services/"+serviceID+"/unregister", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify deleted
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/services/"+serviceID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestConfigBuilder(t *testing.T) {
	cfg, err := config.NewBuilder().
		WithEnv().
		Validate().
		Build()

	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "release", cfg.GinMode)
}
