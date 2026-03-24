package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	appdb "github.com/user/u-status/internal/db"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down>")
	}
	command := os.Args[1]

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://ustatus:ustatus@localhost:5432/ustatus?sslmode=disable"
	}

	db, err := appdb.Open(databaseURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := runMigration(db, command); err != nil {
		log.Fatalf("migration %s failed: %v", command, err)
	}
	log.Printf("migration %s completed successfully", command)
}

func runMigration(db *sql.DB, command string) error {
	goose.SetBaseFS(appdb.EmbedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}

	switch command {
	case "up":
		return goose.Up(db, "migrations")
	case "down":
		return goose.Down(db, "migrations")
	default:
		return fmt.Errorf("unknown command: %s (use 'up' or 'down')", command)
	}
}
