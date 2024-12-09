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
	// Debug logging first for request visibility
	a.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			a.logger.Debug("incoming request",
				zap.String("origin", c.Request().Header.Get(echo.HeaderOrigin)),
				zap.String("method", c.Request().Method),
				zap.String("path", c.Request().URL.Path),
				zap.Any("headers", c.Request().Header),
			)
			return next(c)
		}
	})

	// CORS middleware before other middleware to properly handle preflight and blocked origins
	a.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     a.config.Security.CorsAllowedOrigins,
		AllowMethods:     a.config.Security.CorsAllowedMethods,
		AllowHeaders:     a.config.Security.CorsAllowedHeaders,
		AllowCredentials: true,
		MaxAge:           a.config.Security.CorsMaxAge,
		ExposeHeaders:    []string{"X-Request-Id"},
	}))

	// Recovery middleware
	a.echo.Use(middleware.Recover())

	// Request ID for tracing
	a.echo.Use(middleware.RequestID())

	// Security headers
	a.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "SAMEORIGIN",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))

	// Rate limiting last
	a.echo.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
		rate.Limit(a.config.RateLimit.Rate))))

	a.echo.HTTPErrorHandler = a.customErrorHandler()

	// Log CORS config
	a.logger.Info("CORS configuration",
		zap.Strings("allowed_origins", a.config.Security.CorsAllowedOrigins),
		zap.Strings("allowed_methods", a.config.Security.CorsAllowedMethods),
		zap.Strings("allowed_headers", a.config.Security.CorsAllowedHeaders),
		zap.Int("max_age", a.config.Security.CorsMaxAge),
	)
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
		code := http.StatusInternalServerError
		message := "Internal Server Error"

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			message = fmt.Sprintf("%v", he.Message)
		}

		a.logger.Error("request error",
			zap.Int("status", code),
			zap.String("message", message),
			zap.Error(err),
			zap.String("path", c.Request().URL.Path),
			zap.String("method", c.Request().Method),
			zap.String("origin", c.Request().Header.Get(echo.HeaderOrigin)),
		)

		if !c.Response().Committed {
			if err := c.JSON(code, map[string]interface{}{
				"error":  message,
				"code":   code,
				"path":   c.Request().URL.Path,
				"method": c.Request().Method,
			}); err != nil {
				a.logger.Error("error sending error response", zap.Error(err))
			}
		}
	}
}
