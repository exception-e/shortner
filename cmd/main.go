package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"shortner/internal/config"
	"shortner/internal/handlers"
	"shortner/internal/logger"
	"shortner/internal/service"
	"shortner/internal/storage"
	"shortner/internal/storage/pgStorage"
	"shortner/internal/storage/types"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	var logHandler = slog.NewJSONHandler(os.Stdout, nil)
	logger := logger.New(logHandler)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := initDB(cfg.Storage)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer func() {
		logger.Info("Closing DB connection")
		err := db.Close()
		if err != nil {
			logger.Error("error closing database", "error", err)
		}
	}()

	if err := runMigrations(db, logger); err != nil {
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	var linkStorage types.LinkStorage

	linkStorage, err = pgStorage.NewPostgresStorage(db, logger)
	if err != nil {
		log.Fatal("Failed to create storage %w ", err)
	}
	linkService, err := service.NewShortnerService(linkStorage, logger, cfg.Service.BaseURL)
	if err != nil {
		log.Fatal("Failed to create service %w", err)
	}
	linkHandler := handlers.NewLinkHandler(linkService, logger)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      setupRouter(linkHandler),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("server failed", "error", err)
				os.Exit(1)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("server stopped gracefully")
}

func initDB(cfg storage.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return db, nil
}

func setupRouter(handler *handlers.LinkHandler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/v1", func(r chi.Router) {
		// Создание короткой ссылки
		r.Post("/shorten", handler.CreateShortLink)

		// статистикa (опционально)
		//r.Get("/stats/{shortCode}", linkHandler.GetStats)
	})

	r.Get("/{alias}", handler.Redirect)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	return r
}

// TODO: миграция вне контекста приложения, т.е. её не отменить
//if err := pgStorage.RunMigrations(dsn); err != nil {
//	log.Fatal("failed to run migrations: ", err)
//}

func runMigrations(db *sql.DB, logger *slog.Logger) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("create postgres driver: %w", err)
	}

	src, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		return fmt.Errorf("open migration files: %w", err)
	}

	m, err := migrate.NewWithInstance("file", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrate instance: %w", err)
	}

	logger.Info("applying database migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("no new migrations to apply")
	} else {
		logger.Info("database migrations applied successfully")
	}

	return nil
}
