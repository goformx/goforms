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
		l.logOnStartExecuting(e)
	case *fxevent.OnStartExecuted:
		l.logOnStartExecuted(e)
	case *fxevent.OnStopExecuting:
		l.logOnStopExecuting(e)
	case *fxevent.OnStopExecuted:
		l.logOnStopExecuted(e)
	case *fxevent.Supplied:
		l.logSupplied(e)
	case *fxevent.Provided:
		l.logProvided(e)
	case *fxevent.Decorated:
		l.logDecorated(e)
	case *fxevent.Invoking:
		l.logInvoking(e)
	case *fxevent.Invoked:
		l.logInvoked(e)
	case *fxevent.Started:
		l.logStarted(e)
	case *fxevent.Stopped:
		l.logStopped(e)
	}
}

func (l *FxEventLogger) logOnStartExecuting(e *fxevent.OnStartExecuting) {
	l.Logger.Debug("fx: start executing",
		String("callee", e.FunctionName),
		String("caller", e.CallerName),
	)
}

func (l *FxEventLogger) logOnStartExecuted(e *fxevent.OnStartExecuted) {
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
}

func (l *FxEventLogger) logOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.Logger.Debug("fx: stop executing",
		String("callee", e.FunctionName),
		String("caller", e.CallerName),
	)
}

func (l *FxEventLogger) logOnStopExecuted(e *fxevent.OnStopExecuted) {
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
}

func (l *FxEventLogger) logSupplied(e *fxevent.Supplied) {
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
}

func (l *FxEventLogger) logProvided(e *fxevent.Provided) {
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
}

func (l *FxEventLogger) logDecorated(e *fxevent.Decorated) {
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
}

func (l *FxEventLogger) logInvoking(e *fxevent.Invoking) {
	l.Logger.Debug("fx: invoking",
		String("function", e.FunctionName),
	)
}

func (l *FxEventLogger) logInvoked(e *fxevent.Invoked) {
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
}

func (l *FxEventLogger) logStarted(e *fxevent.Started) {
	if e.Err != nil {
		l.Logger.Error("fx: start failed", Error(e.Err))
	} else {
		l.Logger.Info("fx: started")
	}
}

func (l *FxEventLogger) logStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.Logger.Error("fx: stop failed", Error(e.Err))
	} else {
		l.Logger.Info("fx: stopped")
	}
}
