package pgStorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"shortner/internal/domain"
	"shortner/internal/storage/types"

	"github.com/jackc/pgx/v5/pgconn"
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

// TODO: здесь должен быть доменный объект (/domain/link)??????????
func (store *PgStorage) PutLink(link *domain.Link) (string, error) {
	_, err := store.db.ExecContext(context.Background(),
		"INSERT INTO links (shortlink, original_url) VALUES ($1, $2)",
		link.Alias, link.OriginalUrl)
	if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
		return "", types.ErrAlreadyExists
	}
	return link.Alias, nil
}

// TODO: здесь должен возвращаться доменный объект ?????????????
func (store *PgStorage) GetLink(alias string) (*domain.Link, error) {
	var dto types.LinkDTO

	err := store.db.QueryRowContext(context.Background(),
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
func (store *PgStorage) ValuePresent(link string) (string, bool) {
	return "", false
}
func (store *PgStorage) KeyPresent(shortLink string) bool {
	return false
}

func (store *PgStorage) Close() error {
	return store.db.Close()
}
