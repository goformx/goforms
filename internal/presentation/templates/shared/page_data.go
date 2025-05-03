package shared

import (
	"github.com/a-h/templ"
	"github.com/jonesrussell/goforms/internal/domain/form"
	"github.com/jonesrussell/goforms/internal/domain/form/model"
	"github.com/jonesrussell/goforms/internal/domain/user"
)

type PageData struct {
	Title         string
	User          *user.User
	Forms         []*form.Form
	Form          *form.Form
	Submissions   []*model.FormSubmission
	CSRFToken     string
	IsDevelopment bool
	AssetPath     func(string) string
	Content       templ.Component
}
