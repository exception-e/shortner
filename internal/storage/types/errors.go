package types

import (
	"errors"
)

var (
	ErrNotFound = errors.New("storage: link not found")

	ErrAlreadyExists = errors.New("storage: alias already exists")

	ErrFetchAlias = errors.New("storage: failed to fetch existing link")
)
