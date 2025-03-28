package main

import (
	"fmt"
	"log"
	"net/http"

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
		Addr: cfg.Address,
		Handler: router,
	}

	fmt.Println("Server started")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server")
	}


}