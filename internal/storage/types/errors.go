package types

import (
	"errors"
)

var (
	ErrNotFound = errors.New("storage: link not found")

	ErrAlreadyExists = errors.New("storage: short code already exists")

	ErrInvalidArgument = errors.New("storage: invalid argument")
)
