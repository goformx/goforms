package models

import "errors"

var (
	ErrDuplicate    = errors.New("duplicate subscription")
	ErrInvalidInput = errors.New("invalid input")
)
