package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"flatty-budget/go-api/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var migrations = []string{
	"migrations/002_backfill_monthly_totals.sql",
	"migrations/003_backfill_monthly_averages.sql",
}

func main() {
	cfg := config.LoadConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	for _, path := range migrations {
		migrationSQL, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("failed to read migration file %s: %v", path, err)
		}

		_, err = pool.Exec(ctx, string(migrationSQL))
		if err != nil {
			log.Fatalf("migration %s failed: %v", path, err)
		}

		fmt.Printf("Migration %s completed successfully\n", path)
	}
}
