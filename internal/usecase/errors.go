package usecase

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotFound           = errors.New("not found")
)
