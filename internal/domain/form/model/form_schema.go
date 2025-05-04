package model

import (
	"time"
)

// FormSchema represents a user-defined form schema (for form builder)
type FormSchema struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	JSONSchema any       `json:"json_schema"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
