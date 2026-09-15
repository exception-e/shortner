package service

import (
	"context"
	"shortner/internal/domain"
)

//go:generate mockgen -source=serviceInterface.go -destination=../../mocks/service/mock.go -package=service_mock

type LinkService interface {
	ShortenLink(ctx context.Context, link string) (alias string, retErr error)
	GetOriginalLink(ctx context.Context, alias string) (*domain.Link, error)
}
