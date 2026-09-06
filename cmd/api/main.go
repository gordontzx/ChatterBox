package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ChatterBox/internal/auth"
	"ChatterBox/internal/chat"
	"ChatterBox/internal/config"
	"ChatterBox/internal/db"
	"ChatterBox/internal/message"
	"ChatterBox/internal/middleware"
	"ChatterBox/internal/user"
	"ChatterBox/internal/ws"
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

	// dependencies
	authStore := auth.NewPostgresStore(pool)
	authSvc := auth.NewService(authStore, []byte(cfg.JWTSecret))
	authHandler := auth.NewHandler(authSvc)

	chatStore := chat.NewPostgresStore(pool)
	chatSvc := chat.NewService(chatStore)
	chatHandler := chat.NewHandler(chatSvc)

	msgStore := message.NewPostgresStore(pool)
	msgSvc := message.NewService(msgStore)
	msgHandler := message.NewHandler(msgSvc)

	userStore := user.NewPostgresStore(pool)
	userSvc := user.NewService(userStore)
	userHandler := user.NewHandler(userSvc)

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, []byte(cfg.JWTSecret), msgSvc)

	// Register routes
	requireAuth := middleware.RequireAuth([]byte(cfg.JWTSecret))
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	mux.Handle("POST /chats/direct", requireAuth(http.HandlerFunc(chatHandler.CreateDirect)))
	mux.Handle("GET /chats", requireAuth(http.HandlerFunc(chatHandler.List)))

	mux.Handle("GET /chats/{chatID}/messages", requireAuth(http.HandlerFunc(msgHandler.GetMessages)))

	mux.Handle("GET /users/me", requireAuth(http.HandlerFunc(userHandler.Me)))
	mux.Handle("GET /users/search", requireAuth(http.HandlerFunc(userHandler.SearchByEmail)))

	mux.HandleFunc("GET /ws", wsHandler.ServeWS)

	// Configure HTTP server
	srv := &http.Server{
		Addr: 		  ":" + cfg.Port,
		Handler: 	  mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("Listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Graceful shutdown on Ctrl+C / SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
