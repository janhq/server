package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "jan-server/services/realtime-api/docs/swagger"
	"jan-server/services/realtime-api/internal/config"
	"jan-server/services/realtime-api/internal/domain/session"
	"jan-server/services/realtime-api/internal/infrastructure/auth"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/handlers"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/middlewares"
	"jan-server/services/realtime-api/internal/interfaces/httpserver/routes"
)

// Note: session.Service is used, session.Store is not needed here
// Session state is now managed via LiveKit polling in the store layer

// HTTPServer is the HTTP server for the realtime API.
type HTTPServer struct {
	cfg         *config.Config
	engine      *gin.Engine
	log         zerolog.Logger
	handlerProv *handlers.Provider
	routeProv   *routes.Provider
}

// New creates a new HTTP server.
func New(
	cfg *config.Config,
	log zerolog.Logger,
	sessionService session.Service,
	authValidator *auth.Validator,
) *HTTPServer {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	// Apply middlewares in order
	engine.Use(middlewares.RequestID())
	engine.Use(middlewares.Tracing(cfg.ServiceName))
	engine.Use(middlewares.Metrics())
	engine.Use(middlewares.CORS())
	engine.Use(middlewares.RequestLoggerWithLogger(log))

	// Public routes (no auth)
	registerCoreRoutes(engine, cfg)

	handlerProvider := handlers.NewProvider(sessionService)
	routeProvider := routes.NewProvider(handlerProvider, authValidator)

	routeProvider.Register(engine)

	return &HTTPServer{
		cfg:         cfg,
		engine:      engine,
		log:         log,
		handlerProv: handlerProvider,
		routeProv:   routeProvider,
	}
}

// Run starts the HTTP server and blocks until context is cancelled.
func (s *HTTPServer) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.cfg.Addr(),
		Handler: s.engine,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info().Str("addr", s.cfg.Addr()).Msg("HTTP server listening")
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Error().Err(err).Msg("HTTP server error")
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
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}

func registerCoreRoutes(engine *gin.Engine, cfg *config.Config) {
	engine.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": cfg.ServiceName,
			"status":  "ok",
		})
	})

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	engine.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Prometheus metrics endpoint
	engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger documentation
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
