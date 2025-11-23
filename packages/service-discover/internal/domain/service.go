package domain

import "time"

type ServiceStatus string

const (
	StatusHealthy   ServiceStatus = "healthy"
	StatusUnhealthy ServiceStatus = "unhealthy"
)

type Route struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
}

type Service struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	Host          string            `json:"host"`
	Port          int               `json:"port"`
	Protocol      string            `json:"protocol"`
	BasePath      string            `json:"base_path"`
	Routes        []Route           `json:"routes,omitempty"`
	HealthCheck   string            `json:"health_check"`
	Tags          []string          `json:"tags,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Status        ServiceStatus     `json:"status"`
	LastHeartbeat time.Time         `json:"last_heartbeat"`
	RegisteredAt  time.Time         `json:"registered_at"`
}

type RegisterServiceRequest struct {
	Name        string            `json:"name" binding:"required"`
	Host        string            `json:"host" binding:"required"`
	Port        int               `json:"port" binding:"required,min=1,max=65535"`
	Protocol    string            `json:"protocol,omitempty"`
	BasePath    string            `json:"base_path,omitempty"`
	Routes      []Route           `json:"routes,omitempty"`
	HealthCheck string            `json:"health_check,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type UpdateServiceRequest struct {
	Host        string            `json:"host,omitempty"`
	Port        int               `json:"port,omitempty" binding:"omitempty,min=1,max=65535"`
	Protocol    string            `json:"protocol,omitempty"`
	BasePath    string            `json:"base_path,omitempty"`
	Routes      []Route           `json:"routes,omitempty"`
	HealthCheck string            `json:"health_check,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
