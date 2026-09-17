package service

import "errors"

var (
	ErrNotFound      = errors.New("service: link not found")
	ErrAlreadyExists = errors.New("service: alias already exists")
	ErrInvalidURL    = errors.New("service: invalid url")
	ErrNotSaved      = errors.New("service: failed to process url")
)
