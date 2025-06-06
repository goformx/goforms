// Package main is the entry point for the GoFormX application.
// It sets up the application using dependency injection (via fx),
// configures the server, and manages the application lifecycle.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/goformx/goforms/internal/application/handlers/web"
	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/config"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	formstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/form"
	userstore "github.com/goformx/goforms/internal/infrastructure/persistence/store/user"
	"github.com/goformx/goforms/internal/presentation/view"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

const (
	readTimeout  = 5 * time.Second
	writeTimeout = 10 * time.Second
	idleTimeout  = 120 * time.Second
)

// ShutdownConfig holds configuration for application shutdown
type ShutdownConfig struct {
	Timeout time.Duration `envconfig:"GOFORMS_SHUTDOWN_TIMEOUT" default:"5s"`
}

func setupLogger() (logging.Logger, error) {
	return logging.NewFactory(logging.FactoryConfig{
		AppName:     "goforms",
		Version:     "1.0.0",
		Environment: "development",
		Fields: map[string]any{
			"version": "1.0.0",
		},
	}).CreateLogger()
}

func setupHandlers(
	baseHandler *web.BaseHandler,
	userService user.Service,
	sessionManager *middleware.SessionManager,
	middlewareManager *middleware.Manager,
	cfg *config.Config,
	logger logging.Logger,
	formService form.Service,
) (*web.WebHandler, *web.AuthHandler, *web.FormHandler, error) {
	// Initialize renderer
	renderer := view.NewRenderer(logger)

	webHandler, err := web.NewWebHandler(web.HandlerDeps{
		BaseHandler:       baseHandler,
		UserService:       userService,
		SessionManager:    sessionManager,
		Renderer:          renderer,
		MiddlewareManager: middlewareManager,
		Config:            cfg,
		Logger:            logger,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create web handler: %w", err)
	}

	authHandler, err := web.NewAuthHandler(web.HandlerDeps{
		BaseHandler:       baseHandler,
		UserService:       userService,
		SessionManager:    sessionManager,
		Renderer:          renderer,
		MiddlewareManager: middlewareManager,
		Config:            cfg,
		Logger:            logger,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create auth handler: %w", err)
	}

	formHandler := web.NewFormHandler(web.HandlerDeps{
		BaseHandler:       baseHandler,
		UserService:       userService,
		SessionManager:    sessionManager,
		Renderer:          renderer,
		MiddlewareManager: middlewareManager,
		Config:            cfg,
		Logger:            logger,
	}, formService)

	return webHandler, authHandler, formHandler, nil
}

// main is the entry point of the application.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Load configuration
	cfg, err := config.New()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		return
	}

	// Initialize logger
	logger, err := setupLogger()
	if err != nil {
		log.Printf("Failed to initialize logger: %v", err)
		return
	}

	// Initialize database
	db, err := database.NewDB(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize repositories
	userRepo := userstore.NewStore(db.DB, logger)
	formRepo := formstore.NewStore(db.DB, logger)

	// Initialize services
	userService := user.NewService(userRepo, logger)
	formService := form.NewService(formRepo)

	// Initialize session manager
	sessionManager := middleware.NewSessionManager(logger)

	// Initialize middleware manager
	middlewareManager := middleware.New(&middleware.ManagerConfig{
		Logger:         logger,
		Security:       &cfg.Security,
		UserService:    userService,
		SessionManager: sessionManager,
		Config:         cfg,
	})

	// Initialize base handler
	baseHandler := web.NewBaseHandler(formService, logger)

	// Initialize handlers
	webHandler, authHandler, formHandler, err := setupHandlers(
		baseHandler,
		userService,
		sessionManager,
		middlewareManager,
		cfg,
		logger,
		formService,
	)
	if err != nil {
		logger.Fatal("Failed to create handlers", zap.Error(err))
	}

	// Initialize router
	router := echo.New()
	router.Use(echomiddleware.Logger())
	router.Use(echomiddleware.Recover())
	router.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     cfg.Security.CorsAllowedOrigins,
		AllowMethods:     cfg.Security.CorsAllowedMethods,
		AllowHeaders:     cfg.Security.CorsAllowedHeaders,
		AllowCredentials: cfg.Security.CorsAllowCredentials,
		MaxAge:           cfg.Security.CorsMaxAge,
	}))

	// Register handlers
	webHandler.Register(router)
	authHandler.Register(router)
	formHandler.Register(router)

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Start server in a goroutine
	go func() {
		serverErr := server.ListenAndServe()
		if serverErr != nil && !errors.Is(serverErr, http.ErrServerClosed) {
			logger.Fatal("Failed to start server", zap.Error(serverErr))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	shutdownErr := server.Shutdown(shutdownCtx)
	if shutdownErr != nil {
		logger.Fatal("Failed to shutdown server", zap.Error(shutdownErr))
	}
}
