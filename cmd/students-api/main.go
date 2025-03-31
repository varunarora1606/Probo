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

	"github.com/gin-gonic/gin"
	"github.com/varunarora1606/Booking-App-Go/internal/config"
	"github.com/varunarora1606/Booking-App-Go/internal/database"
	"github.com/varunarora1606/Booking-App-Go/internal/http/handlers/order"
	"github.com/varunarora1606/Booking-App-Go/internal/http/handlers/user"
	"github.com/varunarora1606/Booking-App-Go/internal/models"
)

func main() {
	// Load config
	cfg := config.MustLoad()

	// Db setup
	database.Connect(cfg.DBUrl)
	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		slog.Error("Failed to migrate database", "error", err.Error())
		os.Exit(1) // Exit if migration fails
	}

	// Setup router
	router := gin.Default()
	router.POST("/api/v1/user/signup", user.Signup)
	router.POST("/api/v1/user/login", user.Signin)
	router.POST("/api/v1/user/logout", user.Logout)
	router.POST("/api/v1/order/buy", order.BuyHandler)
	router.POST("/api/v1/order/sell", order.SellHandler)

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
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", err.Error())
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
