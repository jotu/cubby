package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joacim/cubby/internal/api"
	"github.com/joacim/cubby/internal/db"
)

func main() {
	port := envOr("CUBBY_PORT", "8080")
	dbPath := envOr("CUBBY_DB_PATH", "cubby.db")
	uploadDir := envOr("CUBBY_UPLOAD_DIR", "uploads")
	frontendDevURL := envOr("CUBBY_FRONTEND_DEV_URL", "")

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer func() { _ = database.Close() }()

	queries := db.NewQueries(database)
	server := api.NewServerWithFrontendProxy(queries, uploadDir, frontendDevURL)

	handler := api.Chain(server, api.LogRequest, api.CORS)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Cubby listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown: %v", err)
	}
	log.Println("Server stopped")
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
