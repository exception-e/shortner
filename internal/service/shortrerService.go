package service

import (
	"context"
	"fmt"
	"log/slog"
	"shortner/internal/domain"
	"shortner/internal/utils"

	storageTypes "shortner/internal/storage/types"
)

type ShortnerService struct {
	linkStorage storageTypes.LinkStorage
	logger      *slog.Logger
}

func NewShortnerService(linkStorage storageTypes.LinkStorage, logger *slog.Logger) (*ShortnerService, error) {
	componentLogger := logger.With(slog.String("component", "shortnerService"))
	return &ShortnerService{linkStorage: linkStorage, logger: componentLogger}, nil
}

func (s *ShortnerService) ShortenLink(ctx context.Context, link string) (string, error) {
	s.logger.Info("Shortening link", slog.String("link", link))

	alias := utils.EncodeBase62(utils.GetHash(link))
	newLink, err := domain.NewLink(link, alias)
	if err != nil {
		return "", fmt.Errorf("service failed to save url %s: %w", link, err)
	}
	existingAlias, err := s.linkStorage.PutLink(ctx, newLink)
	if err != nil {
		return "", fmt.Errorf("service failed to save url %s: %w", link, err)
	}
	s.logger.Info("Link shortened and saved", slog.String("alias", existingAlias))
	return "http://localhost:8080/" + existingAlias, nil
}

func (s *ShortnerService) GetOriginalLink(ctx context.Context, alias string) (*domain.Link, error) {
	s.logger.Info("Getting original link", slog.String("shortLink", alias))
	link, err := s.linkStorage.GetLink(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get original url for alias %s: %w", alias, err)
	}
	s.logger.Info("Original link for alias", slog.String("alias", link.Alias))
	return link, nil
}
