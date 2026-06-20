package filehost

import "errors"

var (
	ErrFolderNotFound = errors.New("folder not found")
	ErrFileNotFound   = errors.New("file not found")
	ErrForbidden      = errors.New("forbidden")
	ErrInvalidName    = errors.New("invalid name")
	ErrFileTooLarge   = errors.New("file too large")
)
