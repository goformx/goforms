package app

import (
	"html/template"

	"github.com/jmoiron/sqlx"
	"github.com/jonesrussell/goforms/internal/models"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, error) {
	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.InfoLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	return config.Build()
}

func NewEcho() *echo.Echo {
	e := echo.New()

	// Add validator
	e.Validator = NewValidator()

	// Serve static files
	e.Static("/static", "static")

	return e
}

func AsModelsDB(db *sqlx.DB) models.DB {
	return db
}

// NewTemplateProvider creates and returns a template provider
func NewTemplateProvider() *template.Template {
	// First parse the base template
	tmpl := template.Must(template.ParseFiles("static/templates/layout.html"))

	// Then parse all other templates that use the base
	template.Must(tmpl.ParseGlob("static/templates/*.html"))

	return tmpl
}
