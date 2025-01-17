package app

import (
	"html/template"
	"log"

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
	e.File("/", "static/index.html")
	e.File("/contact", "static/contact.html")

	return e
}

func AsModelsDB(db *sqlx.DB) models.DB {
	return db
}

// NewTemplateProvider parses and returns the HTML templates
func NewTemplateProvider() *template.Template {
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	return templates
}
