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
		l.Logger.Debug("OnStart executing", String("callee", e.FunctionName), String("caller", e.CallerName))
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			l.Logger.Error("OnStart failed", String("callee", e.FunctionName), String("caller", e.CallerName), Error(e.Err))
		} else {
			l.Logger.Debug("OnStart executed", String("callee", e.FunctionName), String("caller", e.CallerName), Duration("runtime", e.Runtime))
		}
	case *fxevent.OnStopExecuting:
		l.Logger.Debug("OnStop executing", String("callee", e.FunctionName), String("caller", e.CallerName))
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			l.Logger.Error("OnStop failed", String("callee", e.FunctionName), String("caller", e.CallerName), Error(e.Err))
		} else {
			l.Logger.Debug("OnStop executed", String("callee", e.FunctionName), String("caller", e.CallerName), Duration("runtime", e.Runtime))
		}
	case *fxevent.Supplied:
		if e.Err != nil {
			l.Logger.Error("supplied failed", String("type", e.TypeName), Error(e.Err))
		} else {
			l.Logger.Debug("supplied", String("type", e.TypeName))
		}
	}
}
