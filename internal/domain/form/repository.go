package form

import (
	"errors"
	"sync"
	"time"

	"github.com/jonesrussell/goforms/internal/domain/form/model"
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

// InMemoryFormSchemaRepo is a simple in-memory implementation for demo/testing
// Not safe for concurrent use in production!
type InMemoryFormSchemaRepo struct {
	mu      sync.Mutex
	schemas map[uint]*model.FormSchema
	nextID  uint
}

// NewInMemoryFormSchemaRepo creates a new in-memory form schema repository
func NewInMemoryFormSchemaRepo() *InMemoryFormSchemaRepo {
	return &InMemoryFormSchemaRepo{
		schemas: make(map[uint]*model.FormSchema),
		nextID:  1,
	}
}

// List returns all form schemas
func (r *InMemoryFormSchemaRepo) List() ([]*model.FormSchema, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	result := make([]*model.FormSchema, 0, len(r.schemas))
	for _, s := range r.schemas {
		result = append(result, s)
	}
	return result, nil
}

// Create stores a new form schema
func (r *InMemoryFormSchemaRepo) Create(schema *model.FormSchema) (*model.FormSchema, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	schema.ID = r.nextID
	r.nextID++
	now := time.Now()
	schema.CreatedAt = now
	schema.UpdatedAt = now
	r.schemas[schema.ID] = schema
	return schema, nil
}

// Get retrieves a form schema by ID
func (r *InMemoryFormSchemaRepo) Get(id uint) (*model.FormSchema, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.schemas[id]
	if !ok {
		return nil, ErrFormSchemaNotFound
	}
	return s, nil
}

// Update modifies an existing form schema
func (r *InMemoryFormSchemaRepo) Update(id uint, schema *model.FormSchema) (*model.FormSchema, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.schemas[id]
	if !ok {
		return nil, ErrFormSchemaNotFound
	}
	schema.ID = id
	schema.CreatedAt = existing.CreatedAt
	schema.UpdatedAt = time.Now()
	r.schemas[id] = schema
	return schema, nil
}

// Delete removes a form schema by ID
func (r *InMemoryFormSchemaRepo) Delete(id uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.schemas, id)
	return nil
}
