package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/user/u-status/internal/api"
	"github.com/user/u-status/internal/checker"
	appdb "github.com/user/u-status/internal/db"
)

func main() {
	dbPath := getenv("DATABASE_PATH", "./data/u-status.db")
	jwtSecret := getenv("JWT_SECRET", "change-me-in-production")
	port := getenv("PORT", "8080")
	allowedOrigins := strings.Split(getenv("ALLOWED_ORIGINS", "http://localhost:5173"), ",")

	// Open database
	db, err := appdb.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := appdb.Migrate(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}
	log.Println("database migrations applied")

	// Channels for API ↔ scheduler communication
	notifyCh := make(chan string, 16)
	deleteCh := make(chan string, 16)

	// Start checker scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := checker.NewScheduler(db, notifyCh, deleteCh)
	sched.Start(ctx)

	// Build router
	router := api.NewRouter(db, jwtSecret, allowedOrigins, notifyCh, deleteCh)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
