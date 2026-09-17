//go:build integration

package pgStorage

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"shortner/internal/domain"
	"shortner/internal/storage/types"
	"shortner/internal/tests"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	if _, err := testcontainers.NewDockerProvider(); err != nil {
		println("integration: Docker unavailable, skipping:", err.Error())
		os.Exit(0) // exit 0, чтобы `go test ./...` оставался зелёным
	}

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		panic(err)
	}
	defer pgContainer.Terminate(ctx)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	testDB, err = sql.Open("pgx", connStr)
	if err != nil {
		panic(err)
	}
	defer testDB.Close()

	if err := testDB.PingContext(ctx); err != nil {
		panic(err)
	}

	if err := runMigrations(testDB); err != nil {
		panic(err)
	}

	code := m.Run()
	pgContainer.Terminate(ctx)
	testDB.Close()

	os.Exit(code)
}

func runMigrations(db *sql.DB) error { //TODO с миграциями
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS links (
            id BIGSERIAL PRIMARY KEY,
            short_link TEXT UNIQUE NOT NULL,
            original_url TEXT NOT NULL,
            created_at TIMESTAMPTZ DEFAULT NOW()
        );
        CREATE INDEX IF NOT EXISTS idx_short_link ON links (short_link);
    `)
	return err
}

func TestPgStorage_PutLink(t *testing.T) {
	ctx := t.Context()
	storage, err := NewPostgresStorage(testDB, slog.New(tests.NewTestHandler(t)))
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		link := &domain.Link{
			Alias:       "test123",
			OriginalURL: "https://google.com",
		}

		alias, err := storage.PutLink(ctx, link)
		require.NoError(t, err)
		assert.Equal(t, "test123", alias)
	})

	t.Run("duplicate alias", func(t *testing.T) {
		// Первый раз — ок
		link1 := &domain.Link{Alias: "3fjKsi", OriginalURL: "https://example.com"}
		_, err := storage.PutLink(ctx, link1)
		require.NoError(t, err)

		// Второй раз — ошибка
		link2 := &domain.Link{Alias: "3fjKsi", OriginalURL: "https://another.com"}
		_, err = storage.PutLink(ctx, link2)
		assert.ErrorIs(t, err, types.ErrAlreadyExists)
	})
}

func TestPgStorage_GetLink(t *testing.T) {
	ctx := context.Background()
	storage, err := NewPostgresStorage(testDB, slog.New(tests.NewTestHandler(t)))
	require.NoError(t, err)

	_, err = testDB.Exec(`TRUNCATE TABLE links RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
	// Fixtures
	link := &domain.Link{Alias: "3fjKsi", OriginalURL: "https://example.com"}
	_, err = storage.PutLink(ctx, link)
	require.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		result, err := storage.GetLink(ctx, "3fjKsi")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", result.OriginalURL)
	})

	t.Run("not found", func(t *testing.T) {
		result, err := storage.GetLink(ctx, "nonexistent")
		assert.ErrorIs(t, err, types.ErrNotFound)
		assert.Nil(t, result)
	})
}
