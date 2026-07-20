package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortner/internal/handlers"
	"shortner/internal/service"
	"shortner/internal/storage"
	"syscall"
	"time"
)

func main() {

	storage := storage.NewMapStorage()
	service := service.NewShortnerService(storage)
	handlers := handlers.NewLinkHandler(service)

	srv := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: handlers,
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

	log.Println("Wait for signal")
	<-ctx.Done()
	log.Println("Signal catched")

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer shutdownCtxCancel()

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
