package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/carlosealves2/video-ia/service-discover/internal/config"
	"github.com/carlosealves2/video-ia/service-discover/internal/handler"
	"github.com/carlosealves2/video-ia/service-discover/internal/logger"
	"github.com/carlosealves2/video-ia/service-discover/internal/middleware"
	"github.com/carlosealves2/video-ia/service-discover/internal/repository"
)

type App struct {
	config  *config.Config
	logger  *zap.Logger
	router  *gin.Engine
	repo    repository.ServiceRepository
	handler *handler.ServiceHandler
}

func New(cfg *config.Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) InitLogger() *App {
	log, err := logger.New(a.config.LogLevel)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	a.logger = log
	return a
}

func (a *App) InitRepository() *App {
	a.repo = repository.NewMemoryRepository()
	return a
}

func (a *App) InitHandlers() *App {
	a.handler = handler.NewServiceHandler(a.repo, a.logger)
	return a
}

func (a *App) InitRouter() *App {
	gin.SetMode(a.config.GinMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logging(a.logger))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		services := api.Group("/services")
		{
			services.POST("/register", a.handler.Register)
			services.GET("/list", a.handler.List)
			services.GET("/search", a.handler.Search)
			services.GET("/:id", a.handler.Get)
			services.PUT("/:id/update", a.handler.Update)
			services.DELETE("/:id/unregister", a.handler.Unregister)
			services.PUT("/:id/heartbeat", a.handler.Heartbeat)
		}
	}

	a.router = router
	return a
}

func (a *App) Run() error {
	a.logger.Info("Starting service-discover",
		zap.Int("port", a.config.Port),
		zap.String("log_level", a.config.LogLevel),
	)

	addr := fmt.Sprintf(":%d", a.config.Port)
	return a.router.Run(addr)
}

func (a *App) GetRouter() *gin.Engine {
	return a.router
}

func (a *App) GetLogger() *zap.Logger {
	return a.logger
}
