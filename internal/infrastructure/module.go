package infrastructure

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/dig"
	"go.uber.org/fx"

	"github.com/jonesrussell/goforms/internal/application/handler"
	appmiddleware "github.com/jonesrussell/goforms/internal/application/middleware"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
	h "github.com/jonesrussell/goforms/internal/handlers"
	webhandler "github.com/jonesrussell/goforms/internal/handlers/web"
	wh_auth "github.com/jonesrussell/goforms/internal/handlers/web/auth"
	"github.com/jonesrussell/goforms/internal/infrastructure/config"
	"github.com/jonesrussell/goforms/internal/infrastructure/database"
	"github.com/jonesrussell/goforms/internal/infrastructure/logging"
	"github.com/jonesrussell/goforms/internal/infrastructure/server"
	"github.com/jonesrussell/goforms/internal/infrastructure/store"
	formstore "github.com/jonesrussell/goforms/internal/infrastructure/store/form"
	"github.com/jonesrussell/goforms/internal/presentation/handlers"
	"github.com/jonesrussell/goforms/internal/presentation/services"
	"github.com/jonesrussell/goforms/internal/presentation/view"
)

const (
	// MinSecretLength is the minimum length required for security secrets
	MinSecretLength = 32
)

// Stores groups all database store providers.
// This struct is used with fx.Out to provide multiple stores
// to the fx container in a single provider function.
type Stores struct {
	fx.Out

	UserStore user.Store
	FormStore form.Store
}

// CoreParams contains core infrastructure dependencies that are commonly needed by handlers.
// These are typically required for basic handler functionality like logging and rendering.
type CoreParams struct {
	fx.In
	Logger   logging.Logger
	Renderer *view.Renderer
	Config   *config.Config
}

// ServiceParams contains all service dependencies that handlers might need.
// This separation makes it easier to manage service dependencies and allows for
// more granular dependency injection.
type ServiceParams struct {
	fx.In
	UserService user.Service
	FormService form.Service
}

// ServiceContainer holds all service instances
type ServiceContainer struct {
	PageDataService *services.PageDataService
	FormOperations  *services.FormOperations
	TemplateService *services.TemplateService
	ResponseBuilder *services.ResponseBuilder
}

// AnnotateHandler is a helper function that simplifies the creation of handler providers.
// It wraps the common fx.Provide and fx.Annotate pattern used for handlers.
func AnnotateHandler(fn any) fx.Option {
	return fx.Provide(
		fx.Annotate(
			fn,
			fx.As(new(h.Handler)),
			fx.ResultTags(`group:"handlers"`),
		),
	)
}

// validateDatabaseConfig validates the database configuration
func validateDatabaseConfig(cfg *config.DatabaseConfig) error {
	if cfg.MaxOpenConns <= 0 {
		return errors.New("database max open connections must be a positive number")
	}
	if cfg.MaxIdleConns <= 0 {
		return errors.New("database max idle connections must be a positive number")
	}
	if cfg.ConnMaxLifetime <= 0 {
		return errors.New("database connection max lifetime must be a positive duration")
	}
	return nil
}

// validateSecurityConfig validates the security configuration
func validateSecurityConfig(cfg *config.SecurityConfig) error {
	if len(cfg.JWTSecret) < MinSecretLength {
		return errors.New("JWT secret must be at least 32 characters long")
	}
	return nil
}

// validateServerConfig validates the server configuration
func validateServerConfig(cfg *config.ServerConfig) error {
	if cfg.Port <= 0 {
		return errors.New("server port must be a positive number")
	}
	if cfg.ReadTimeout <= 0 {
		return errors.New("server read timeout must be a positive duration")
	}
	if cfg.WriteTimeout <= 0 {
		return errors.New("server write timeout must be a positive duration")
	}
	return nil
}

// validateConfig checks if the configuration is valid
func validateConfig(cfg *config.Config, logger logging.Logger) error {
	logger.Info("validateConfig: Starting configuration validation...")
	logger.Info("validateConfig: Database config",
		logging.Int("MaxOpenConns", cfg.Database.MaxOpenConns),
		logging.Int("MaxIdleConns", cfg.Database.MaxIdleConns),
		logging.Duration("ConnMaxLifetime", cfg.Database.ConnMaxLifetime))
	logger.Info("validateConfig: JWT Secret length",
		logging.Int("length", len(cfg.Security.JWTSecret)))

	var validationErrors []string

	// Validate database configuration
	logger.Info("validateConfig: Validating database configuration...")
	if dbErr := validateDatabaseConfig(&cfg.Database); dbErr != nil {
		logger.Error("validateConfig: Database validation failed",
			logging.Error(dbErr))
		validationErrors = append(validationErrors, dbErr.Error())
	}

	// Validate security configuration
	logger.Info("validateConfig: Validating security configuration...")
	if secErr := validateSecurityConfig(&cfg.Security); secErr != nil {
		logger.Error("validateConfig: Security validation failed",
			logging.Error(secErr))
		validationErrors = append(validationErrors, secErr.Error())
	}

	// Validate server configuration
	logger.Info("validateConfig: Validating server configuration...")
	if srvErr := validateServerConfig(&cfg.Server); srvErr != nil {
		logger.Error("validateConfig: Server validation failed",
			logging.Error(srvErr))
		validationErrors = append(validationErrors, srvErr.Error())
	}

	if len(validationErrors) > 0 {
		logger.Error("validateConfig: Validation failed",
			logging.StringField("errors", strings.Join(validationErrors, "; ")))
		return fmt.Errorf("configuration validation failed: %s", strings.Join(validationErrors, "; "))
	}

	logger.Info("validateConfig: Configuration validation successful")
	return nil
}

// Module represents the infrastructure module
type Module struct {
	app            *fx.App
	container      *dig.Container
	config         *config.Config
	logger         logging.Logger
	db             *sql.DB
	formService    form.Service
	userService    user.Service
	authMiddleware *appmiddleware.CookieAuthMiddleware
	services       *ServiceContainer
	handler        *handlers.Handler
}

// NewModule creates a new infrastructure module
func NewModule(
	app *fx.App,
	container *dig.Container,
	appConfig *config.Config,
	logger logging.Logger,
	db *sql.DB,
	formService form.Service,
	userService user.Service,
	authMiddleware *appmiddleware.CookieAuthMiddleware,
) *Module {
	m := &Module{
		app:            app,
		container:      container,
		config:         appConfig,
		logger:         logger,
		db:             db,
		formService:    formService,
		userService:    userService,
		authMiddleware: authMiddleware,
	}

	m.initializeServices()
	m.initializeHandlers()

	return m
}

// InfrastructureModule provides core infrastructure dependencies.
var InfrastructureModule = fx.Options(
	fx.Provide(
		func(logger logging.Logger) (*config.Config, error) {
			logger.Info("InfrastructureModule: Starting configuration loading...")
			cfg, cfgErr := config.New(logger)
			if cfgErr != nil {
				logger.Error("InfrastructureModule: Error loading configuration",
					logging.Error(cfgErr))
				return nil, fmt.Errorf("failed to load configuration: %w", cfgErr)
			}

			logger.Info("InfrastructureModule: Configuration loaded, starting validation...")
			if validationErr := validateConfig(cfg, logger); validationErr != nil {
				logger.Error("InfrastructureModule: Validation failed",
					logging.Error(validationErr))
				return nil, validationErr
			}

			logger.Info("InfrastructureModule: Configuration validated successfully")
			return cfg, nil
		},
		database.NewDB,
	),
)

// StoreModule provides all database store implementations.
var StoreModule = fx.Options(
	fx.Provide(NewStores),
)

// HandlerModule provides all HTTP handlers for the application.
var HandlerModule = fx.Options(
	// Static file handler (must be first)
	AnnotateHandler(func(core CoreParams) (h.Handler, error) {
		handler := handler.NewStaticHandler(core.Logger, core.Config)
		if handler == nil {
			return nil, errors.New("failed to create static handler")
		}
		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "StaticHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
	// Web handlers
	AnnotateHandler(func(core CoreParams, middlewareManager *appmiddleware.Manager) (h.Handler, error) {
		handler := webhandler.NewHomeHandler(core.Logger, core.Renderer)
		if handler == nil {
			return nil, fmt.Errorf("failed to create home handler: renderer=%T", core.Renderer)
		}
		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "HomeHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
	AnnotateHandler(func(core CoreParams, middlewareManager *appmiddleware.Manager) (h.Handler, error) {
		handler := wh_auth.NewWebLoginHandler(core.Logger, core.Renderer)
		if handler == nil {
			return nil, fmt.Errorf("failed to create web login handler: renderer=%T", core.Renderer)
		}
		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "WebLoginHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
	AnnotateHandler(
		func(
			core CoreParams,
			services ServiceParams,
			middlewareManager *appmiddleware.Manager,
		) (h.Handler, error) {
			handler, err := handler.NewWebHandler(
				core.Logger,
				core.Renderer,
				handler.WithMiddlewareManager(middlewareManager),
				handler.WithConfig(core.Config),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create web handler: %w", err)
			}
			core.Logger.Debug(
				"registered handler",
				logging.StringField("handler_name", "WebHandler"),
				logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
				logging.StringField("operation", "handler_registration"),
			)
			return handler, nil
		},
	),
	// Auth handler
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		handler := handler.NewAuthHandler(core.Logger, handler.WithUserService(services.UserService))
		if handler == nil {
			return nil, fmt.Errorf("failed to create auth handler: user_service=%T", services.UserService)
		}
		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "AuthHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
	// Dashboard handler
	AnnotateHandler(func(core CoreParams, services ServiceParams) (h.Handler, error) {
		authMiddleware := appmiddleware.NewCookieAuthMiddleware(services.UserService, core.Logger)
		baseHandler := handlers.NewBaseHandler(
			authMiddleware,
			services.FormService,
			core.Logger,
		)
		handler, err := handlers.NewHandler(services.UserService, services.FormService, core.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create dashboard handler: %w", err)
		}

		// Initialize the handler with the base handler
		handler.DashboardHandler.Base = baseHandler
		handler.FormHandler.Base = baseHandler
		handler.SubmissionHandler.Base = baseHandler
		handler.SchemaHandler.Base = baseHandler

		core.Logger.Debug("registered handler",
			logging.StringField("handler_name", "DashboardHandler"),
			logging.StringField("handler_type", fmt.Sprintf("%T", handler)),
			logging.StringField("operation", "handler_registration"),
		)
		return handler, nil
	}),
)

// ServerModule provides the HTTP server setup.
var ServerModule = fx.Options(
	fx.Provide(server.New),
)

// RootModule combines all infrastructure-level modules into a single module.
var RootModule = fx.Options(
	InfrastructureModule,
	StoreModule,
	ServerModule,
	HandlerModule,
)

// wrapCreator creates a type-safe wrapper for store creation functions
func wrapCreator[T any](
	creator func(*database.Database, logging.Logger) T,
) func(*database.Database, logging.Logger) any {
	return func(db *database.Database, logger logging.Logger) any {
		return creator(db, logger)
	}
}

// wrapAssigner creates a type-safe wrapper for store assignment functions
func wrapAssigner[T any](assigner func(*Stores, T)) func(*Stores, any) {
	return func(s *Stores, instance any) {
		if s == nil {
			panic(errors.New("database connection is nil"))
		}
		typedInstance, ok := instance.(T)
		if !ok {
			panic(errors.New("invalid instance type"))
		}
		assigner(s, typedInstance)
	}
}

// NewStores creates all database stores.
// This function is responsible for initializing all database stores
// and providing them to the fx container.
func NewStores(db *database.Database, logger logging.Logger) (Stores, error) {
	if db == nil {
		logger.Error("database connection is nil",
			logging.StringField("operation", "store_initialization"),
			logging.StringField("error_type", "nil_database"),
		)
		return Stores{}, errors.New("database connection is nil")
	}

	startTime := time.Now()

	// Map of store creators
	storeCreators := map[string]struct {
		create func(*database.Database, logging.Logger) any
		assign func(*Stores, any)
	}{
		"user": {
			create: wrapCreator(store.NewUserStore),
			assign: wrapAssigner(func(s *Stores, v user.Store) { s.UserStore = v }),
		},
		"form": {
			create: wrapCreator(formstore.NewStore),
			assign: wrapAssigner(func(s *Stores, v form.Store) { s.FormStore = v }),
		},
	}

	var stores Stores
	var wg sync.WaitGroup
	var mu sync.Mutex
	failedStores := make(chan string, len(storeCreators))
	results := make(chan struct {
		name     string
		instance any
	}, len(storeCreators))

	// Initialize stores concurrently
	for name, creator := range storeCreators {
		wg.Add(1)
		go func(name string, creator struct {
			create func(*database.Database, logging.Logger) any
			assign func(*Stores, any)
		}) {
			defer wg.Done()

			// Create store instance
			instance := creator.create(db, logger)
			if instance == nil {
				logger.Error("store creation failed",
					logging.StringField("store_type", name),
					logging.StringField("operation", "store_initialization"),
					logging.StringField("error_type", "nil_instance"),
				)
				failedStores <- name
				return
			}

			results <- struct {
				name     string
				instance any
			}{name, instance}
		}(name, creator)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(failedStores)
	close(results)

	// Collect failed stores
	var failedStoreNames []string
	for name := range failedStores {
		failedStoreNames = append(failedStoreNames, name)
	}

	// Handle initialization errors
	if len(failedStoreNames) > 0 {
		return Stores{}, fmt.Errorf("failed to initialize the following stores: %s", strings.Join(failedStoreNames, ", "))
	}

	// Assign successful stores
	for result := range results {
		mu.Lock()
		storeCreators[result.name].assign(&stores, result.instance)
		mu.Unlock()
	}

	// Log successful initialization metrics
	logger.Info("all database stores initialized successfully",
		logging.StringField("operation", "store_initialization"),
		logging.DurationField("init_duration", time.Since(startTime)),
		logging.IntField("total_stores_initialized", len(storeCreators)),
	)

	return stores, nil
}

// initializeServices initializes all services
func (m *Module) initializeServices() {
	m.services = &ServiceContainer{
		PageDataService: services.NewPageDataService(m.logger),
		FormOperations:  services.NewFormOperations(m.formService, m.logger),
		TemplateService: services.NewTemplateService(m.logger),
		ResponseBuilder: services.NewResponseBuilder(m.logger),
	}
}

// initializeHandlers initializes all handlers
func (m *Module) initializeHandlers() {
	// Create base handler
	baseHandler := handlers.NewBaseHandler(m.authMiddleware, m.formService, m.logger)

	// Create feature handlers
	dashboardHandler := handlers.NewDashboardHandler(m.formService, m.logger, baseHandler)
	formHandler := handlers.NewFormHandler(m.formService, m.services.FormOperations, m.logger, baseHandler)
	submissionHandler := handlers.NewSubmissionHandler(m.formService, m.logger, baseHandler)
	schemaHandler := handlers.NewSchemaHandler(m.formService, m.logger, baseHandler)

	// Create main handler
	mainHandler, err := handlers.NewHandler(m.userService, m.formService, m.logger)
	if err != nil {
		m.logger.Error("failed to create handler", logging.Error(err))
		return
	}

	// Set the handlers
	mainHandler.DashboardHandler = dashboardHandler
	mainHandler.FormHandler = formHandler
	mainHandler.SubmissionHandler = submissionHandler
	mainHandler.SchemaHandler = schemaHandler

	m.handler = mainHandler
}
