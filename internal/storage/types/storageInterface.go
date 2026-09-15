package types

import (
	"context"
	"shortner/internal/domain"
)

//go:generate mockgen -source=storageInterface.go -destination=../../../mocks/storage/mock.go -package=storage_mock

// TODO: godocs
// TODO: ctx
type LinkStorage interface {
	PutLink(ctx context.Context, link *domain.Link) (string, error)
	GetLink(ctx context.Context, alias string) (*domain.Link, error)
}
