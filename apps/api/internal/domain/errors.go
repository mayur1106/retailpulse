package domain

import "errors"

var (
	ErrConflict          = errors.New("resource already exists")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrForbidden         = errors.New("forbidden")
	ErrNotFound          = errors.New("resource not found")
	ErrValidation        = errors.New("validation failed")
	ErrConfiguration     = errors.New("configuration error")
)
