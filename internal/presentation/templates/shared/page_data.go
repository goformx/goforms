package shared

import "github.com/jonesrussell/goforms/internal/domain/user"

type PageData struct {
	Title     string
	CSRFToken string
	User      *user.User
}
