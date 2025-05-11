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
		StringField("callee", e.FunctionName),
		StringField("caller", e.CallerName),
	)
}

func (l *FxEventLogger) logOnStartExecuted(e *fxevent.OnStartExecuted) {
	if e.Err != nil {
		l.Logger.Error("fx: start error",
			StringField("callee", e.FunctionName),
			StringField("caller", e.CallerName),
			ErrorField("error", e.Err),
		)
	} else {
		l.Logger.Debug("fx: started",
			StringField("callee", e.FunctionName),
			StringField("caller", e.CallerName),
			DurationField("runtime", e.Runtime),
		)
	}
}

func (l *FxEventLogger) logOnStopExecuting(e *fxevent.OnStopExecuting) {
	l.Logger.Debug("fx: stop executing",
		StringField("callee", e.FunctionName),
		StringField("caller", e.CallerName),
	)
}

func (l *FxEventLogger) logOnStopExecuted(e *fxevent.OnStopExecuted) {
	if e.Err != nil {
		l.Logger.Error("fx: stop error",
			StringField("callee", e.FunctionName),
			StringField("caller", e.CallerName),
			ErrorField("error", e.Err),
		)
	} else {
		l.Logger.Debug("fx: stopped",
			StringField("callee", e.FunctionName),
			StringField("caller", e.CallerName),
			DurationField("runtime", e.Runtime),
		)
	}
}

func (l *FxEventLogger) logSupplied(e *fxevent.Supplied) {
	if e.Err != nil {
		l.Logger.Error("fx: supplied error",
			StringField("type", e.TypeName),
			ErrorField("error", e.Err),
		)
	} else {
		l.Logger.Debug("fx: supplied",
			StringField("type", e.TypeName),
		)
	}
}

func (l *FxEventLogger) logProvided(e *fxevent.Provided) {
	for _, rtype := range e.OutputTypeNames {
		l.Logger.Debug("fx: provided",
			StringField("constructor", e.ConstructorName),
			StringField("type", rtype),
		)
	}
	if e.Err != nil {
		l.Logger.Error("fx: error providing",
			StringField("constructor", e.ConstructorName),
			ErrorField("error", e.Err),
		)
	}
}

func (l *FxEventLogger) logDecorated(e *fxevent.Decorated) {
	for _, rtype := range e.OutputTypeNames {
		l.Logger.Debug("fx: decorated",
			StringField("decorator", e.DecoratorName),
			StringField("type", rtype),
		)
	}
	if e.Err != nil {
		l.Logger.Error("fx: error decorating",
			StringField("decorator", e.DecoratorName),
			ErrorField("error", e.Err),
		)
	}
}

func (l *FxEventLogger) logInvoking(e *fxevent.Invoking) {
	l.Logger.Debug("fx: invoking",
		StringField("function", e.FunctionName),
	)
}

func (l *FxEventLogger) logInvoked(e *fxevent.Invoked) {
	if e.Err != nil {
		l.Logger.Error("fx: invoke failed",
			StringField("function", e.FunctionName),
			ErrorField("error", e.Err),
		)
	} else {
		l.Logger.Debug("fx: invoked",
			StringField("function", e.FunctionName),
		)
	}
}

func (l *FxEventLogger) logStarted(e *fxevent.Started) {
	if e.Err != nil {
		l.Logger.Error("fx: start failed", ErrorField("error", e.Err))
	} else {
		l.Logger.Info("fx: started")
	}
}

func (l *FxEventLogger) logStopped(e *fxevent.Stopped) {
	if e.Err != nil {
		l.Logger.Error("fx: stop failed", ErrorField("error", e.Err))
	} else {
		l.Logger.Info("fx: stopped")
	}
}
