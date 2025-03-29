package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/varunarora1606/Booking-App-Go/internal/config"
)

func main() {
	// Load config
	cfg := config.MustLoad()

	// Db setup

	// Setup router
	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!!"))
	})

	// Setup server
	server := http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	slog.Info("Server started on", slog.String("Address", cfg.Address))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Println("Server started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed{
			log.Fatal("Failed to start server")
		}
	}()

	<-done

	slog.Info("Received shutdown signal, shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down server", slog.String("error", err.Error()))
	} else {
		slog.Info("Server shutdown successfully")
	}

}
