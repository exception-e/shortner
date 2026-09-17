package types

import (
	"context"
	"shortner/internal/domain"
)

//go:generate mockgen -source=storage_interface.go -destination=../../../mocks/storage/mock.go -package=storage_mock

// TODO: godocs
type LinkStorage interface {
	PutLink(ctx context.Context, link *domain.Link) (string, error)
	GetLink(ctx context.Context, alias string) (*domain.Link, error)
	FindExistingAlias(ctx context.Context, link *domain.Link) (string, error)
}
