package shared

import (
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/user"
)

type PageData struct {
	Title         string
	User          *user.User
	Forms         []*form.Form
	Form          *form.Form
	CSRFToken     string
	IsDevelopment bool
}
