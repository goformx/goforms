package logging

import (
	"go.uber.org/fx/fxevent"
)

// FxEventLogger implements fxevent.Logger interface using our Logger
type FxEventLogger struct {
	Logger Logger
}

// LogEvent logs fx lifecycle events
func (l *FxEventLogger) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		l.Logger.Debug("fx: start executing",
			String("callee", e.FunctionName),
			String("caller", e.CallerName),
		)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.Logger.Error("fx: start error",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Error(e.Err),
			)
		} else {
			l.Logger.Debug("fx: started",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.OnStopExecuting:
		l.Logger.Debug("fx: stop executing",
			String("callee", e.FunctionName),
			String("caller", e.CallerName),
		)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.Logger.Error("fx: stop error",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Error(e.Err),
			)
		} else {
			l.Logger.Debug("fx: stopped",
				String("callee", e.FunctionName),
				String("caller", e.CallerName),
				Duration("runtime", e.Runtime),
			)
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.Logger.Error("fx: supplied error",
				String("type", e.TypeName),
				Error(e.Err),
			)
		} else {
			l.Logger.Debug("fx: supplied",
				String("type", e.TypeName),
			)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			l.Logger.Debug("fx: provided",
				String("constructor", e.ConstructorName),
				String("type", rtype),
			)
		}
		if e.Err != nil {
			l.Logger.Error("fx: error providing",
				String("constructor", e.ConstructorName),
				Error(e.Err),
			)
		}
	case *fxevent.Decorated:
		for _, rtype := range e.OutputTypeNames {
			l.Logger.Debug("fx: decorated",
				String("decorator", e.DecoratorName),
				String("type", rtype),
			)
		}
		if e.Err != nil {
			l.Logger.Error("fx: error decorating",
				String("decorator", e.DecoratorName),
				Error(e.Err),
			)
		}
	case *fxevent.Invoking:
		l.Logger.Debug("fx: invoking",
			String("function", e.FunctionName),
		)
	case *fxevent.Invoked:
		if e.Err != nil {
			l.Logger.Error("fx: invoke failed",
				String("function", e.FunctionName),
				Error(e.Err),
			)
		} else {
			l.Logger.Debug("fx: invoked",
				String("function", e.FunctionName),
			)
		}
	case *fxevent.Started:
		if e.Err != nil {
			l.Logger.Error("fx: start failed", Error(e.Err))
		} else {
			l.Logger.Info("fx: started")
		}
	case *fxevent.Stopped:
		if e.Err != nil {
			l.Logger.Error("fx: stop failed", Error(e.Err))
		} else {
			l.Logger.Info("fx: stopped")
		}
	}
}
