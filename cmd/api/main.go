package main

import (
	"context"
	"log"
	"net/http"

	"ChatterBox/internal/auth"
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

	// Dependencies
	store := auth.NewPostgresStore(pool)
	authSvc := auth.NewService(store, []byte(cfg.JWTSecret))
	authHandler := auth.NewHandler(authSvc)

	// Register routes
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	log.Printf("Listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":" + cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
