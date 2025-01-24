package logging

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"

	"github.com/jonesrussell/goforms/internal/infrastructure/config"
)

// Result bundles logger-related dependencies for injection.
type Result struct {
	fx.Out

	Logger Logger
	Fxlog  fxevent.Logger `group:"fx.logging"`
}

// New constructs a new logger and its fx event logger.
func New(cfg *config.Config) Result {
	logger := NewLogger(&cfg.App)

	// Create an fx event logger that uses our logger
	fxLogger := &FxEventLogger{logger}

	return Result{
		Logger: logger,
		Fxlog:  fxLogger,
	}
}

// Module provides logging dependencies.
//
//nolint:gochecknoglobals // This is an fx module definition, which is meant to be global
var Module = fx.Options(
	fx.Provide(New),
)

// FxEventLogger adapts our Logger to fx's logging interface
type FxEventLogger struct {
	Logger Logger
}

// LogEvent implements fxevent.Logger interface
func (l *FxEventLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.Logger.Info("OnStart hook executing",
			String("callee", e.FunctionName),
			String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.Logger.Error("OnStart hook failed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Error(e.Err),
			)
		} else {
			l.Logger.Info("OnStart hook executed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.OnStopExecuting:
		l.Logger.Info("OnStop hook executing",
			String("callee", e.FunctionName),
			String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.Logger.Error("OnStop hook failed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Error(e.Err),
			)
		} else {
			l.Logger.Info("OnStop hook executed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.Logger.Error("supplied",
				String("type", e.TypeName),
				Error(e.Err),
			)
		} else {
			l.Logger.Info("supplied",
				String("type", e.TypeName),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			l.Logger.Error("provided",
				String("constructor", e.ConstructorName),
				Error(e.Err),
			)
		} else {
			l.Logger.Info("provided",
				String("constructor", e.ConstructorName),
				String("type", e.OutputTypeNames[0]),
			)
		}
	case *fxevent.Invoking:
		l.Logger.Info("invoking",
			String("function", e.FunctionName),
		)
	case *fxevent.Started:
		if e.Err != nil {
			l.Logger.Error("start failed", Error(e.Err))
		} else {
			l.Logger.Info("started")
		}
	case *fxevent.Stopped:
		if e.Err != nil {
			l.Logger.Error("stop failed", Error(e.Err))
		} else {
			l.Logger.Info("stopped")
		}
	}
}
