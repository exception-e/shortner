package pgStorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"shortner/internal/domain"
	"shortner/internal/storage/types"
	"shortner/internal/utils"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PgStorage struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewPostgresStorage(dsn string, logger *slog.Logger) (*PgStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	} // TODO: fmt.Errorf("sql.Open: %w")???????
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to postgres db %w", err)
	}

	logger.Info("Connected to pg database")

	return &PgStorage{
		db:     db,
		logger: logger,
	}, nil
}

func (store *PgStorage) PutLink(ctx context.Context, link *domain.Link) (string, error) {
	var existingAlias string
	alias := link.Alias
	const attemptCounter = 10

	for i := 0; i < attemptCounter; i++ {
		if i > 0 {
			alias = utils.AddSalt(link.Alias, i)
		}
		err := store.db.QueryRowContext(ctx,
			"INSERT INTO links (short_link, original_url) "+
				"VALUES ($1, $2)"+
				"ON CONFLICT(short_link) DO NOTHING"+
				"RETURNING short_link",
			alias, link.OriginalUrl).Scan(&existingAlias)
		if err == nil {
			return existingAlias, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("failed to insert link: %w", err)
		}

		if errors.Is(err, sql.ErrNoRows) {
			err = store.db.QueryRowContext(ctx, "SELECT short_link FROM links WHERE original_url = $1", link.OriginalUrl).Scan(&existingAlias)
			if err == nil {
				return existingAlias, nil
			}
			if !errors.Is(err, sql.ErrNoRows) {
				return "", fmt.Errorf("failed to insert link: %w", err)
			}
		}
	}
	return "", fmt.Errorf("failed to insert link after %d attempts", attemptCounter)
}

func (store *PgStorage) GetLink(ctx context.Context, alias string) (*domain.Link, error) {
	var dto types.LinkDTO

	err := store.db.QueryRowContext(ctx,
		"SELECT id, short_code, original_url, created_at FROM links WHERE shortlink == $1",
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
			OriginalUrl: dto.OriginalURL,
			CreatedAt:   dto.CreatedAt},
		nil
}

func (store *PgStorage) Close() error {
	return store.db.Close()
}
