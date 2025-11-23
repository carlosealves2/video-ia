package repository

import (
	"errors"
	"sync"

	"github.com/carlosealves2/video-ia/service-discover/internal/domain"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceAlreadyExists = errors.New("service already exists")
)

type ServiceRepository interface {
	Create(service *domain.Service) error
	GetByID(id string) (*domain.Service, error)
	GetAll() []*domain.Service
	Update(service *domain.Service) error
	Delete(id string) error
	Exists(id string) bool
}

type MemoryRepository struct {
	services map[string]*domain.Service
	mu       sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		services: make(map[string]*domain.Service),
	}
}

func (r *MemoryRepository) Create(service *domain.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[service.ID]; exists {
		return ErrServiceAlreadyExists
	}

	r.services[service.ID] = service
	return nil
}

func (r *MemoryRepository) GetByID(id string) (*domain.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	service, exists := r.services[id]
	if !exists {
		return nil, ErrServiceNotFound
	}

	return service, nil
}

func (r *MemoryRepository) GetAll() []*domain.Service {
	r.mu.RLock()
	defer r.mu.RUnlock()

	services := make([]*domain.Service, 0, len(r.services))
	for _, service := range r.services {
		services = append(services, service)
	}

	return services
}

func (r *MemoryRepository) Update(service *domain.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[service.ID]; !exists {
		return ErrServiceNotFound
	}

	r.services[service.ID] = service
	return nil
}

func (r *MemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.services[id]; !exists {
		return ErrServiceNotFound
	}

	delete(r.services, id)
	return nil
}

func (r *MemoryRepository) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.services[id]
	return exists
}
