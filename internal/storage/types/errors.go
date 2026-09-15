package types

import (
	"errors"
)

var (
	ErrNotFound = errors.New("storage: link not found")

	ErrAlreadyExists = errors.New("storage: alias already exists")

	ErrInvalidArgument = errors.New("storage: invalid argument")
)
