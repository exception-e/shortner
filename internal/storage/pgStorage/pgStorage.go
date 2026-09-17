package pgStorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"shortner/internal/domain"
	"shortner/internal/storage/types"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgStorage struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPostgresStorage(db *sql.DB, logger *slog.Logger) (*PgStorage, error) {
	componentLogger := logger.With(slog.String("component", "postgres-storage"))
	return &PgStorage{
		db:     db,
		logger: componentLogger,
	}, nil
}

func (store *PgStorage) PutLink(ctx context.Context, link *domain.Link) (string, error) {
	var existingAlias string
	alias := link.Alias
	err := store.db.QueryRowContext(ctx,
		"INSERT INTO links (short_link, original_url) "+
			"VALUES ($1, $2)"+
			"ON CONFLICT(short_link) DO NOTHING "+
			"RETURNING short_link",
		alias, link.OriginalURL).Scan(&existingAlias)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			//if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return "", types.ErrAlreadyExists
		}
		return "", fmt.Errorf("failed to insert link: %w", err)
	}
	return existingAlias, nil
}

func (store *PgStorage) FindExistingAlias(ctx context.Context, link *domain.Link) (string, error) {
	var existingAlias string
	err := store.db.QueryRowContext(ctx, "SELECT short_link FROM links WHERE original_url = $1", link.OriginalURL).Scan(&existingAlias)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", types.ErrFetchAlias
		}
		return "", types.ErrNotFound
	}
	return existingAlias, nil
}

func (store *PgStorage) GetLink(ctx context.Context, alias string) (*domain.Link, error) {
	var dto types.LinkDTO

	err := store.db.QueryRowContext(ctx,
		"SELECT id, short_link, original_url, created_at FROM links WHERE short_link = $1",
		alias).
		Scan(&dto.ID, &dto.Alias, &dto.OriginalURL, &dto.CreatedAt)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, types.ErrNotFound
		}
		return nil, fmt.Errorf("failed to fulfill query: %w", err)
	}

	return &domain.Link{
			Id:          dto.ID,
			Alias:       dto.Alias,
			OriginalURL: dto.OriginalURL,
			CreatedAt:   dto.CreatedAt},
		nil
}

func (store *PgStorage) Close() error {
	return store.db.Close()
}
