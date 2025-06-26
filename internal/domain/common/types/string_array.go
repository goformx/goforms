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
	data, err := json.Marshal(*a)
	if err != nil {
		return nil, fmt.Errorf("marshal string array: %w", err)
	}
	return data, nil
}

// Scan implements the sql.Scanner interface.
func (a *StringArray) Scan(src any) error {
	if src == nil {
		*a = []string{}
		return nil
	}

	switch v := src.(type) {
	case []byte:
		if err := json.Unmarshal(v, a); err != nil {
			return fmt.Errorf("unmarshal string array from bytes: %w", err)
		}
		return nil
	case string:
		if err := json.Unmarshal([]byte(v), a); err != nil {
			return fmt.Errorf("unmarshal string array from string: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("cannot scan %T into StringArray", src)
	}
}

// MarshalJSON implements json.Marshaler.
func (a *StringArray) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal([]string(*a))
	if err != nil {
		return nil, fmt.Errorf("marshal string array to JSON: %w", err)
	}
	return data, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *StringArray) UnmarshalJSON(data []byte) error {
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return fmt.Errorf("unmarshal JSON to string array: %w", err)
	}
	*a = arr
	return nil
}
