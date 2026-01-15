package framework

import (
	"errors"

	"go.uber.org/fx"

	"github.com/goformx/goforms/internal/domain/common/events"
	"github.com/goformx/goforms/internal/domain/form"
	"github.com/goformx/goforms/internal/domain/user"
	"github.com/goformx/goforms/internal/infrastructure/database"
	"github.com/goformx/goforms/internal/infrastructure/logging"
	formstore "github.com/goformx/goforms/internal/infrastructure/repository/form"
	formsubmissionstore "github.com/goformx/goforms/internal/infrastructure/repository/form/submission"
	userstore "github.com/goformx/goforms/internal/infrastructure/repository/user"
)

func domainModule() fx.Option {
	return fx.Module(
		"domain",
		fx.Provide(
			provideUserService,
			provideFormService,
			provideStores,
		),
	)
}

type storeResult struct {
	fx.Out
	UserRepository           user.Repository
	FormRepository           form.Repository
	FormSubmissionRepository form.SubmissionRepository
}

func provideStores(db database.DB, logger logging.Logger) (storeResult, error) {
	if db == nil {
		return storeResult{}, errors.New("database connection is required")
	}

	if logger == nil {
		return storeResult{}, errors.New("logger is required")
	}

	userRepo := userstore.NewStore(db, logger)
	formRepo := formstore.NewStore(db, logger)
	formSubmissionRepo := formsubmissionstore.NewStore(db, logger)

	if userRepo == nil || formRepo == nil || formSubmissionRepo == nil {
		logger.Error("failed to create repository",
			"operation", "repository_initialization",
			"repository_type", "user/form/submission",
			"error_type", "nil_repository",
		)

		return storeResult{}, errors.New("failed to create repository: one or more repositories are nil")
	}

	return storeResult{
		UserRepository:           userRepo,
		FormRepository:           formRepo,
		FormSubmissionRepository: formSubmissionRepo,
	}, nil
}

func provideUserService(repo user.Repository, logger logging.Logger) (user.Service, error) {
	if repo == nil {
		return nil, errors.New("user repository is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return user.NewService(repo, logger), nil
}

func provideFormService(
	repo form.Repository,
	eventBus events.EventBus,
	logger logging.Logger,
) (form.Service, error) {
	if repo == nil {
		return nil, errors.New("form repository is required")
	}

	if eventBus == nil {
		return nil, errors.New("event bus is required")
	}

	if logger == nil {
		return nil, errors.New("logger is required")
	}

	return form.NewService(repo, eventBus, logger), nil
}
