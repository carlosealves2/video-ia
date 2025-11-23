package servicediscovery

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:8080")
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8080", client.baseURL)
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient("http://localhost:8080",
		WithTimeout(5*time.Second),
		WithRetries(5),
		WithRetryDelay(2*time.Second),
	)
	assert.NotNil(t, client)
	assert.Equal(t, 5*time.Second, client.options.timeout)
	assert.Equal(t, 5, client.options.retries)
	assert.Equal(t, 2*time.Second, client.options.retryDelay)
}

func TestRegister(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/register", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "test-service", req.Name)
		assert.Equal(t, "localhost", req.Host)
		assert.Equal(t, 3000, req.Port)

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(&Service{
			ID:            "test-id",
			Name:          req.Name,
			Host:          req.Host,
			Port:          req.Port,
			Status:        StatusHealthy,
			RegisteredAt:  time.Now(),
			LastHeartbeat: time.Now(),
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service, err := client.Register(context.Background(), &RegisterRequest{
		Name: "test-service",
		Host: "localhost",
		Port: 3000,
	})

	require.NoError(t, err)
	assert.Equal(t, "test-id", service.ID)
	assert.Equal(t, "test-service", service.Name)
}

func TestAutoRegister(t *testing.T) {
	_ = os.Setenv("SERVICE_NAME", "auto-service")
	_ = os.Setenv("PORT", "4000")
	_ = os.Setenv("SERVICE_TAGS", "api,v1")
	defer func() {
		_ = os.Unsetenv("SERVICE_NAME")
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("SERVICE_TAGS")
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "auto-service", req.Name)
		assert.Equal(t, 4000, req.Port)
		assert.Equal(t, []string{"api", "v1"}, req.Tags)

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(&Service{
			ID:   "auto-id",
			Name: req.Name,
			Port: req.Port,
			Tags: req.Tags,
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service, err := client.AutoRegister(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "auto-service", service.Name)
	assert.Equal(t, 4000, service.Port)
}

func TestAutoRegisterWithOverrides(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "override-service", req.Name)
		assert.Equal(t, 5000, req.Port)

		w.WriteHeader(http.StatusCreated)
		err = json.NewEncoder(w).Encode(&Service{
			ID:   "override-id",
			Name: req.Name,
			Port: req.Port,
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service, err := client.AutoRegister(context.Background(),
		WithName("override-service"),
		WithPort(5000),
	)

	require.NoError(t, err)
	assert.Equal(t, "override-service", service.Name)
}

func TestList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/list", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&ListResponse{
			Services: []*Service{
				{ID: "1", Name: "service-1"},
				{ID: "2", Name: "service-2"},
			},
			Count: 2,
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	services, err := client.List(context.Background())

	require.NoError(t, err)
	assert.Len(t, services, 2)
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/test-id", r.URL.Path)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&Service{
			ID:   "test-id",
			Name: "test-service",
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	service, err := client.Get(context.Background(), "test-id")

	require.NoError(t, err)
	assert.Equal(t, "test-id", service.ID)
}

func TestGetNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode(map[string]string{"error": "service not found"})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Get(context.Background(), "not-found")

	assert.ErrorIs(t, err, ErrServiceNotFound)
}

func TestHeartbeat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/test-id/heartbeat", r.URL.Path)
		assert.Equal(t, http.MethodPut, r.Method)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&HeartbeatResponse{
			Message:       "heartbeat received",
			LastHeartbeat: time.Now(),
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Heartbeat(context.Background(), "test-id")

	require.NoError(t, err)
}

func TestUnregister(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/test-id/unregister", r.URL.Path)
		assert.Equal(t, http.MethodDelete, r.Method)

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]string{"message": "service unregistered successfully"})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.Unregister(context.Background(), "test-id")

	require.NoError(t, err)
}

func TestSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/services/search", r.URL.Path)
		assert.Equal(t, "/users", r.URL.Query().Get("route"))

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(&ListResponse{
			Services: []*Service{
				{ID: "1", Name: "user-service"},
			},
			Count: 1,
		})
		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	services, err := client.Search(context.Background(), "/users", "", "")

	require.NoError(t, err)
	assert.Len(t, services, 1)
}
