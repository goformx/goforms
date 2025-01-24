package logging

import (
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// Result bundles logger-related dependencies for injection.
type Result struct {
	fx.Out

	Logger Logger
	Fxlog  fxevent.Logger `group:"fx.logging"`
}

// New constructs a new logger and its fx event logger.
func New() Result {
	logger := NewLogger()

	// Create an fx event logger that uses our logger
	fxLogger := &fxEventLogger{logger}

	return Result{
		Logger: logger,
		Fxlog:  fxLogger,
	}
}

// Module provides logging dependencies.
var Module = fx.Options(
	fx.Provide(New),
)

// fxEventLogger adapts our Logger to fx's logging interface
type fxEventLogger struct {
	log Logger
}

// LogEvent implements fxevent.Logger interface
func (l *fxEventLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.log.Info("OnStart hook executing",
			String("callee", e.FunctionName),
			String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.log.Error("OnStart hook failed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Error(e.Err),
			)
		} else {
			l.log.Info("OnStart hook executed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.OnStopExecuting:
		l.log.Info("OnStop hook executing",
			String("callee", e.FunctionName),
			String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.log.Error("OnStop hook failed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Error(e.Err),
			)
		} else {
			l.log.Info("OnStop hook executed",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.log.Error("supplied",
				String("type", e.TypeName),
				Error(e.Err),
			)
		} else {
			l.log.Info("supplied",
				String("type", e.TypeName),
			)
		}
	case *fxevent.Provided:
		if e.Err != nil {
			l.log.Error("provided",
				String("constructor", e.ConstructorName),
				Error(e.Err),
			)
		} else {
			l.log.Info("provided",
				String("constructor", e.ConstructorName),
				String("type", e.OutputTypeNames[0]),
			)
		}
	case *fxevent.Invoking:
		l.log.Info("invoking",
			String("function", e.FunctionName),
		)
	case *fxevent.Started:
		if e.Err != nil {
			l.log.Error("start failed", Error(e.Err))
		} else {
			l.log.Info("started")
		}
	case *fxevent.Stopped:
		if e.Err != nil {
			l.log.Error("stop failed", Error(e.Err))
		} else {
			l.log.Info("stopped")
		}
	}
}
