package form

import (
	"errors"

	"github.com/goformx/goforms/internal/domain/form/model"
)

var (
	// ErrFormSchemaNotFound is returned when a form schema cannot be found
	ErrFormSchemaNotFound = errors.New("form schema not found")
)

// SchemaRepository defines the interface for form schema storage
type SchemaRepository interface {
	List() ([]*model.FormSchema, error)
	Create(schema *model.FormSchema) (*model.FormSchema, error)
	Get(id uint) (*model.FormSchema, error)
	Update(id uint, schema *model.FormSchema) (*model.FormSchema, error)
	Delete(id uint) error
}
