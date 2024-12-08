package main

import (
	"github.com/jonesrussell/goforms/internal/app"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		app.Module,
	).Run()
}
