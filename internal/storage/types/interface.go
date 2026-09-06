package types

import (
	"context"
	"shortner/internal/domain"
)

// TODO: godocs
// TODO: ctx
type LinkStorage interface {
	PutLink(ctx context.Context, link *domain.Link) (string, error)
	GetLink(ctx context.Context, alias string) (*domain.Link, error)
}
