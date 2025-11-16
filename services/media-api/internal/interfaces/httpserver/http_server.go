package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	mediaapidocs "jan-server/services/media-api/docs/swagger"
	"jan-server/services/media-api/internal/config"
	domain "jan-server/services/media-api/internal/domain/media"
	"jan-server/services/media-api/internal/infrastructure/auth"
	"jan-server/services/media-api/internal/interfaces/httpserver/handlers"
	v1 "jan-server/services/media-api/internal/interfaces/httpserver/routes/v1"
)

// HttpServer wraps the gin engine with graceful shutdown helpers.
type HttpServer struct {
	cfg    *config.Config
	engine *gin.Engine
	log    zerolog.Logger
	auth   *auth.Validator
}

// New constructs the HTTP server with default middleware and routes.
func New(cfg *config.Config, log zerolog.Logger, mediaService *domain.Service, authValidator *auth.Validator) *HttpServer {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	mediaapidocs.SwaggerInfo.BasePath = "/"

	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger())

	handlerProvider := handlers.NewProvider(cfg, mediaService, log)
	routeProvider := v1.NewRoutes(handlerProvider, cfg)
	if authValidator != nil {
		engine.Use(authValidator.Middleware())
	}
	registerCoreRoutes(engine, cfg, routeProvider, authValidator)

	return &HttpServer{
		cfg:    cfg,
		engine: engine,
		log:    log,
		auth:   authValidator,
	}
}

// Run starts the HTTP listener and handles graceful shutdown via context cancellation.
func (s *HttpServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.cfg.Addr(),
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info().Str("addr", s.cfg.Addr()).Msg("media-api HTTP server listening")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		s.log.Info().Msg("context cancelled, shutting down HTTP server")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.ShutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func registerCoreRoutes(engine *gin.Engine, cfg *config.Config, routes *v1.Routes, authValidator *auth.Validator) {
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": cfg.ServiceName, "status": "ok"})
	})
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
	engine.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	engine.GET("/health/auth", func(c *gin.Context) {
		if authValidator == nil || authValidator.Ready() {
			c.JSON(http.StatusOK, gin.H{"status": "ready"})
			return
		}
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "initializing"})
	})
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.Register(engine.Group("/"))
}
