package framework

import (
	"go.uber.org/fx"

	"github.com/labstack/echo/v4"

	"github.com/goformx/goforms/internal/application/middleware"
	"github.com/goformx/goforms/internal/application/middleware/request"
	"github.com/goformx/goforms/internal/application/response"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	"github.com/goformx/goforms/internal/infrastructure/sanitization"
)

func applicationModule() fx.Option {
	return fx.Module(
		"application",
		fx.Provide(
			provideRequestUtils,
			provideErrorHandler,
			provideRecoveryMiddleware,
		),
	)
}

func provideRequestUtils(sanitizer sanitization.ServiceInterface) *request.Utils {
	return request.NewUtils(sanitizer)
}

func provideErrorHandler(
	logger logging.Logger,
	sanitizer sanitization.ServiceInterface,
) response.ErrorHandlerInterface {
	return response.NewErrorHandler(logger, sanitizer)
}

func provideRecoveryMiddleware(
	logger logging.Logger,
	sanitizer sanitization.ServiceInterface,
) echo.MiddlewareFunc {
	return middleware.Recovery(logger, sanitizer)
}
