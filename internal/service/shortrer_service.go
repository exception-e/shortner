package service

import (
	"context"
	"fmt"
	"log/slog"
	"shortner/internal/domain"
	"shortner/internal/utils"

	storageTypes "shortner/internal/storage/types"

	"github.com/pkg/errors"
)

type ShortnerService struct {
	linkStorage storageTypes.LinkStorage
	logger      *slog.Logger
	baseURL     string
}

func NewShortnerService(linkStorage storageTypes.LinkStorage, logger *slog.Logger, baseUrl string) (*ShortnerService, error) {
	componentLogger := logger.With(slog.String("component", "shortnerService"))

	return &ShortnerService{linkStorage: linkStorage, logger: componentLogger, baseURL: baseUrl}, nil
}

func (s *ShortnerService) ShortenLink(ctx context.Context, link string) (string, error) {
	s.logger.InfoContext(ctx, "Shortening link", slog.String("link", link))

	alias := utils.EncodeBase62(utils.GetHash(link))
	newLink, err := domain.NewLink(link, alias)
	if err != nil {
		return "", fmt.Errorf("invalid url %s: %w: %w", link, ErrInvalidURL, err)
	}

	existingAlias, err := s.linkStorage.PutLink(ctx, newLink)
	if err != nil {
		if errors.Is(err, storageTypes.ErrAlreadyExists) { // unique_violation
			return "", fmt.Errorf("service failed to save url %s: %w: %w", link, ErrAlreadyExists, err)
		}
		return "", fmt.Errorf("db error %s: %w: %w", link, ErrNotSaved, err)
	}
	s.logger.InfoContext(ctx, "Link shortened and saved", slog.String("alias", existingAlias))

	return s.baseURL + existingAlias, nil
}

func (s *ShortnerService) GetOriginalLink(ctx context.Context, alias string) (*domain.Link, error) {
	s.logger.InfoContext(ctx, "Getting original link", slog.String("shortLink", alias))
	link, err := s.linkStorage.GetLink(ctx, alias)
	if err != nil {
		return nil, fmt.Errorf("service: failed to get original url for alias %s: %w: %w", alias, ErrNotFound, err)
	}
	s.logger.InfoContext(ctx, "Original link for alias", slog.String("alias", link.Alias))

	return link, nil
}
