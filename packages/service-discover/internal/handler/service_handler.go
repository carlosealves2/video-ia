package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/carlosealves2/video-ia/service-discover/internal/domain"
	"github.com/carlosealves2/video-ia/service-discover/internal/repository"
)

type ServiceHandler struct {
	repo   repository.ServiceRepository
	logger *zap.Logger
}

func NewServiceHandler(repo repository.ServiceRepository, logger *zap.Logger) *ServiceHandler {
	return &ServiceHandler{
		repo:   repo,
		logger: logger,
	}
}

func (h *ServiceHandler) Register(c *gin.Context) {
	var req domain.RegisterServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind register request",
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	protocol := req.Protocol
	if protocol == "" {
		protocol = "http"
	}

	healthCheck := req.HealthCheck
	if healthCheck == "" {
		healthCheck = "/health"
	}

	service := &domain.Service{
		ID:            uuid.New().String(),
		Name:          req.Name,
		Host:          req.Host,
		Port:          req.Port,
		Protocol:      protocol,
		BasePath:      req.BasePath,
		Routes:        req.Routes,
		HealthCheck:   healthCheck,
		Tags:          req.Tags,
		Metadata:      req.Metadata,
		Status:        domain.StatusHealthy,
		LastHeartbeat: time.Now(),
		RegisteredAt:  time.Now(),
	}

	if err := h.repo.Create(service); err != nil {
		h.logger.Error("Failed to register service",
			zap.String("service_name", req.Name),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Service registered successfully",
		zap.String("service_id", service.ID),
		zap.String("service_name", service.Name),
		zap.String("host", service.Host),
		zap.Int("port", service.Port),
		zap.String("protocol", service.Protocol),
		zap.String("base_path", service.BasePath),
		zap.Int("routes_count", len(service.Routes)),
	)

	c.JSON(http.StatusCreated, service)
}

func (h *ServiceHandler) List(c *gin.Context) {
	services := h.repo.GetAll()

	h.logger.Info("Listed all services",
		zap.Int("count", len(services)),
	)

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

func (h *ServiceHandler) Get(c *gin.Context) {
	id := c.Param("id")

	service, err := h.repo.GetByID(id)
	if err != nil {
		h.logger.Warn("Service not found",
			zap.String("service_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	h.logger.Info("Retrieved service",
		zap.String("service_id", id),
		zap.String("service_name", service.Name),
	)

	c.JSON(http.StatusOK, service)
}

func (h *ServiceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	service, err := h.repo.GetByID(id)
	if err != nil {
		h.logger.Warn("Service not found for update",
			zap.String("service_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	var req domain.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind update request",
			zap.String("service_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Host != "" {
		service.Host = req.Host
	}
	if req.Port != 0 {
		service.Port = req.Port
	}
	if req.Protocol != "" {
		service.Protocol = req.Protocol
	}
	if req.BasePath != "" {
		service.BasePath = req.BasePath
	}
	if req.Routes != nil {
		service.Routes = req.Routes
	}
	if req.HealthCheck != "" {
		service.HealthCheck = req.HealthCheck
	}
	if req.Tags != nil {
		service.Tags = req.Tags
	}
	if req.Metadata != nil {
		service.Metadata = req.Metadata
	}

	if err := h.repo.Update(service); err != nil {
		h.logger.Error("Failed to update service",
			zap.String("service_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Service updated successfully",
		zap.String("service_id", id),
		zap.String("service_name", service.Name),
	)

	c.JSON(http.StatusOK, service)
}

func (h *ServiceHandler) Unregister(c *gin.Context) {
	id := c.Param("id")

	service, err := h.repo.GetByID(id)
	if err != nil {
		h.logger.Warn("Service not found for unregister",
			zap.String("service_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	serviceName := service.Name

	if err := h.repo.Delete(id); err != nil {
		h.logger.Error("Failed to unregister service",
			zap.String("service_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Service unregistered successfully",
		zap.String("service_id", id),
		zap.String("service_name", serviceName),
	)

	c.JSON(http.StatusOK, gin.H{"message": "service unregistered successfully"})
}

func (h *ServiceHandler) Heartbeat(c *gin.Context) {
	id := c.Param("id")

	service, err := h.repo.GetByID(id)
	if err != nil {
		h.logger.Warn("Service not found for heartbeat",
			zap.String("service_id", id),
		)
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	service.LastHeartbeat = time.Now()
	service.Status = domain.StatusHealthy

	if err := h.repo.Update(service); err != nil {
		h.logger.Error("Failed to update heartbeat",
			zap.String("service_id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Heartbeat received",
		zap.String("service_id", id),
		zap.String("service_name", service.Name),
		zap.Time("last_heartbeat", service.LastHeartbeat),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":        "heartbeat received",
		"last_heartbeat": service.LastHeartbeat,
	})
}

func (h *ServiceHandler) Search(c *gin.Context) {
	route := c.Query("route")
	name := c.Query("name")
	tag := c.Query("tag")

	services := h.repo.GetAll()
	var results []*domain.Service

	for _, svc := range services {
		if route != "" {
			found := false
			for _, r := range svc.Routes {
				if strings.HasPrefix(r.Path, route) || r.Path == route {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		if name != "" && !strings.Contains(strings.ToLower(svc.Name), strings.ToLower(name)) {
			continue
		}

		if tag != "" {
			found := false
			for _, t := range svc.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		results = append(results, svc)
	}

	h.logger.Info("Search completed",
		zap.String("route", route),
		zap.String("name", name),
		zap.String("tag", tag),
		zap.Int("results", len(results)),
	)

	c.JSON(http.StatusOK, gin.H{
		"services": results,
		"count":    len(results),
	})
}
