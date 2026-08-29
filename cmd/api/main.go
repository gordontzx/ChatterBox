package main

import (
	"context"
	"log"
	// "log"
	// "net/http"
	// "os"
	// "os/signal"
	// "syscall"
	// "time"

	// "ChatterBox/internal/auth"
	"ChatterBox/internal/config"
	"ChatterBox/internal/db"
)

func main() {
	cfg := config.Load()

	// Connect to database
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v", err)
	}
	log.Println("Connected to database")
}
