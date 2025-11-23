package repository

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/carlosealves2/video-ia/service-discover/internal/domain"
)

func createTestService(id, name string) *domain.Service {
	return &domain.Service{
		ID:            id,
		Name:          name,
		Host:          "localhost",
		Port:          8080,
		Protocol:      "http",
		Status:        domain.StatusHealthy,
		LastHeartbeat: time.Now(),
		RegisteredAt:  time.Now(),
	}
}

func TestNewMemoryRepository(t *testing.T) {
	repo := NewMemoryRepository()
	assert.NotNil(t, repo)
	assert.NotNil(t, repo.services)
}

func TestCreate(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")

	err := repo.Create(service)

	require.NoError(t, err)
	assert.True(t, repo.Exists("1"))
}

func TestCreateDuplicate(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")

	err := repo.Create(service)
	require.NoError(t, err)

	err = repo.Create(service)
	assert.ErrorIs(t, err, ErrServiceAlreadyExists)
}

func TestGetByID(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")
	_ = repo.Create(service)

	result, err := repo.GetByID("1")

	require.NoError(t, err)
	assert.Equal(t, "1", result.ID)
	assert.Equal(t, "test-service", result.Name)
}

func TestGetByIDNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	_, err := repo.GetByID("not-found")

	assert.ErrorIs(t, err, ErrServiceNotFound)
}

func TestGetAll(t *testing.T) {
	repo := NewMemoryRepository()
	_ = repo.Create(createTestService("1", "service-1"))
	_ = repo.Create(createTestService("2", "service-2"))
	_ = repo.Create(createTestService("3", "service-3"))

	services := repo.GetAll()

	assert.Len(t, services, 3)
}

func TestGetAllEmpty(t *testing.T) {
	repo := NewMemoryRepository()

	services := repo.GetAll()

	assert.Empty(t, services)
}

func TestUpdate(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")
	_ = repo.Create(service)

	service.Name = "updated-service"
	service.Port = 9090
	err := repo.Update(service)

	require.NoError(t, err)

	result, _ := repo.GetByID("1")
	assert.Equal(t, "updated-service", result.Name)
	assert.Equal(t, 9090, result.Port)
}

func TestUpdateNotFound(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("not-found", "test-service")

	err := repo.Update(service)

	assert.ErrorIs(t, err, ErrServiceNotFound)
}

func TestDelete(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")
	_ = repo.Create(service)

	err := repo.Delete("1")

	require.NoError(t, err)
	assert.False(t, repo.Exists("1"))
}

func TestDeleteNotFound(t *testing.T) {
	repo := NewMemoryRepository()

	err := repo.Delete("not-found")

	assert.ErrorIs(t, err, ErrServiceNotFound)
}

func TestExists(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")
	_ = repo.Create(service)

	assert.True(t, repo.Exists("1"))
	assert.False(t, repo.Exists("2"))
}

func TestConcurrentAccess(t *testing.T) {
	repo := NewMemoryRepository()
	var wg sync.WaitGroup

	// Concurrent creates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			service := createTestService(string(rune('A'+id)), "service")
			_ = repo.Create(service)
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = repo.GetAll()
		}()
	}

	wg.Wait()
}

func TestConcurrentReadWrite(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")
	_ = repo.Create(service)

	var wg sync.WaitGroup

	// Concurrent updates and reads
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			s := createTestService("1", "updated")
			_ = repo.Update(s)
		}()

		go func() {
			defer wg.Done()
			_, _ = repo.GetByID("1")
		}()
	}

	wg.Wait()
}

func TestServiceWithRoutes(t *testing.T) {
	repo := NewMemoryRepository()
	service := &domain.Service{
		ID:       "1",
		Name:     "api-service",
		Host:     "localhost",
		Port:     8080,
		Protocol: "http",
		BasePath: "/api/v1",
		Routes: []domain.Route{
			{Path: "/users", Methods: []string{"GET", "POST"}},
			{Path: "/orders", Methods: []string{"GET"}},
		},
		HealthCheck:   "/health",
		Tags:          []string{"api", "v1"},
		Metadata:      map[string]string{"version": "1.0.0"},
		Status:        domain.StatusHealthy,
		LastHeartbeat: time.Now(),
		RegisteredAt:  time.Now(),
	}

	err := repo.Create(service)
	require.NoError(t, err)

	result, err := repo.GetByID("1")
	require.NoError(t, err)

	assert.Equal(t, "/api/v1", result.BasePath)
	assert.Len(t, result.Routes, 2)
	assert.Equal(t, "/users", result.Routes[0].Path)
	assert.Equal(t, []string{"GET", "POST"}, result.Routes[0].Methods)
	assert.Equal(t, []string{"api", "v1"}, result.Tags)
	assert.Equal(t, "1.0.0", result.Metadata["version"])
}

func TestDeleteAndRecreate(t *testing.T) {
	repo := NewMemoryRepository()
	service := createTestService("1", "test-service")

	_ = repo.Create(service)
	_ = repo.Delete("1")

	err := repo.Create(service)
	require.NoError(t, err)
	assert.True(t, repo.Exists("1"))
}
