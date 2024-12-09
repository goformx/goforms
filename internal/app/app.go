package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/jonesrussell/goforms/internal/config"
	"github.com/jonesrussell/goforms/internal/handlers"
)

type App struct {
	logger      *zap.Logger
	echo        *echo.Echo
	config      *config.Config
	handlers    *handlers.SubscriptionHandler
	serverError chan error
}

func NewApp(
	lc fx.Lifecycle,
	logger *zap.Logger,
	echo *echo.Echo,
	cfg *config.Config,
	handler *handlers.SubscriptionHandler,
	healthHandler *handlers.HealthHandler,
) *App {
	app := &App{
		logger:      logger,
		echo:        echo,
		config:      cfg,
		handlers:    handler,
		serverError: make(chan error, 1),
	}

	app.setupMiddleware()
	app.registerHandlers()
	healthHandler.Register(app.echo)

	lc.Append(fx.Hook{
		OnStart: app.start,
		OnStop:  app.stop,
	})

	return app
}

func (a *App) setupMiddleware() {
	// Recovery should be first
	a.echo.Use(middleware.Recover())

	// Logging middleware
	a.echo.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","remote_ip":"${remote_ip}",` +
			`"method":"${method}","uri":"${uri}","status":${status},` +
			`"latency":${latency},"latency_human":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		Output: os.Stdout,
	}))

	// Request ID for tracing
	a.echo.Use(middleware.RequestID())

	// Security middleware
	a.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))

	// CORS middleware
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     a.config.Security.CorsAllowedOrigins,
		AllowMethods:     a.config.Security.CorsAllowedMethods,
		AllowHeaders:     a.config.Security.CorsAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           a.config.Security.CorsMaxAge,
	}))

	// Rate limiting should be last
	a.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(a.config.RateLimit.Rate))))

	a.echo.HTTPErrorHandler = a.customErrorHandler()
}

func (a *App) registerHandlers() {
	a.handlers.Register(a.echo)
}

func (a *App) start(ctx context.Context) error {
	// Always bind to all interfaces
	bindAddress := fmt.Sprintf(":%d", a.config.Server.Port)
	a.logger.Info("starting server",
		zap.String("bind", bindAddress),
		zap.String("host", a.config.Server.Host))

	go func() {
		if err := a.echo.Start(bindAddress); err != nil && err != http.ErrServerClosed {
			a.serverError <- err
			a.logger.Error("server error", zap.Error(err))
		}
	}()

	// Monitor for server errors
	go func() {
		select {
		case err := <-a.serverError:
			a.logger.Error("server error detected", zap.Error(err))
			os.Exit(1)
		case <-ctx.Done():
			return
		}
	}()

	return nil
}

func (a *App) stop(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("shutdown error", zap.Error(err))
		return err
	}

	// Close database connections
	// Clean up other resources
	return nil
}

func (a *App) customErrorHandler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := 500
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = he.Message.(string)
		}

		a.logger.Error("request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
		)

		if !c.Response().Committed {
			if err := c.JSON(code, map[string]string{
				"error": message,
			}); err != nil {
				a.logger.Error("error sending error response", zap.Error(err))
			}
		}
	}
}

// Module returns the application fx module
func Module() fx.Option {
	return NewModule()
}
