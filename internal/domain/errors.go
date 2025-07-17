package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInternalServerError = errors.New("internal server error")
)
