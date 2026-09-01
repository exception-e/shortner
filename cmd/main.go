package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"shortner/internal/handlers"
	"shortner/internal/service"
	"shortner/internal/storage/pgStorage"
	"syscall"
	"time"
)

func main() {

	var logHandler slog.Handler
	logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(logHandler)
	dsn := "postgres://tiffany:password@localhost:5432/linkdb?sslmode=disable"

	// TODO: миграция вне контекста приложения, т.е. её не отменить
	if err := pgStorage.RunMigrations(dsn); err != nil {
		log.Fatal("failed to run migrations: ", err)
	}

	logger.Info("migrations applied")

	pgStorage, err := pgStorage.NewPostgresStorage(dsn, logger)
	if err != nil {
		log.Fatal("Failed to create storage %w ", err)
	}
	defer pgStorage.Close()

	shortnerService := service.NewShortnerService(pgStorage, logger)
	linkHandler := handlers.NewLinkHandler(shortnerService, logger)

	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: linkHandler,
	}

	ctx, ctxCancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer ctxCancel()

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatalln("Server err: ", err)
			}
		}
	}()

	logger.Info("Server started", slog.String("addr", srv.Addr))

	log.Println("Wait for signal")
	<-ctx.Done()
	log.Println("Signal caught")
	logger.Info("Signal caught, shutting down server")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCtxCancel()

	// TODO: здесь должен быть и db.Close с тем же контекстом
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("srv.Shutdown: %v", err)
		return
	}

	// http.HandleFunc("/", handlers.Handler)
	// err := http.ListenAndServe(":8080", nil)

	// if err != nil {
	// 	fmt.Printf("Server error: %v", err)
	// }
}
