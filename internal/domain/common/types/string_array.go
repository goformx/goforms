// Package types provides custom domain types and utilities for handling
// specialized data structures used throughout the application.
package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// StringArray is a custom type for JSON array columns.
type StringArray []string

// Value implements the driver.Valuer interface.
func (a *StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "[]", nil
	}
	return json.Marshal(*a)
}

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = []string{}
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, a)
	case string:
		return json.Unmarshal([]byte(v), a)
	default:
		return fmt.Errorf("cannot scan %T into StringArray", src)
	}
}

// MarshalJSON implements json.Marshaler.
func (a *StringArray) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string(*a))
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *StringArray) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}
