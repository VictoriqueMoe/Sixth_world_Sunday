package event

import "errors"

var (
	ErrEventNotFound = errors.New("event not found")
	ErrForbidden     = errors.New("forbidden")
	ErrInvalidInput  = errors.New("invalid input")
)
