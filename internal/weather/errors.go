package weather

import "errors"

var (
	ErrLocationNotFound = errors.New("weather location not found")
	ErrForbidden        = errors.New("forbidden")
	ErrInvalidInput     = errors.New("invalid input")
	ErrTooMany          = errors.New("too many saved locations")
)
