package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"time"

	"flatty-budget/go-api/internal/config"

	"github.com/pressly/goose/v3"

	_ "github.com/jackc/pgx/v5/stdlib"

	"flatty-budget/go-api/migrations"
)

func main() {
	status := flag.Bool("status", false, "show migration status and exit")
	down := flag.Bool("down", false, "roll back the most recent migration")

	flag.Parse()

	cfg := config.LoadConfig()

	db, err := sql.Open("pgx", cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer db.Close()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)

	err = db.PingContext(pingCtx)

	cancelPing()

	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	ctx := context.Background()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	switch {
	case *status && *down:
		log.Fatal("-status and -down are mutually exclusive")
	case *status:
		if err := goose.StatusContext(ctx, db, "."); err != nil {
			log.Fatalf("status failed: %v", err)
		}
	case *down:
		if err := goose.DownContext(ctx, db, "."); err != nil {
			log.Fatalf("down failed: %v", err)
		}
	default:
		if err := goose.UpContext(ctx, db, "."); err != nil {
			log.Fatalf("up failed: %v", err)
		}
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		log.Fatalf("failed to get db version: %v", err)
	}
	log.Printf("current migration version: %d", version)
}
