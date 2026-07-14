// @title			GO API
// @version		1.0
// @description	TBD
// @BasePath		/api
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"flatty-budget/go-api/internal/config"
	kafkaclient "flatty-budget/go-api/internal/kafka"
	expensesrepo "flatty-budget/go-api/repos/expenses"
	expensesservice "flatty-budget/go-api/services/expenses"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.LoadConfig()
	fmt.Println(3)

	pool, err := pgxpool.New(context.Background(), cfg.DatabaseUrl)

	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}

	defer pool.Close()

	brokers := strings.Split(cfg.KafkaBrockers, ",")

	producer := kafkaclient.NewProducer(brokers, cfg.KafkaTopic)
	defer producer.Close()

	statsUpdater := kafkaclient.NewStatsRepoFromPool(pool)
	consumer := kafkaclient.NewConsumer(brokers, cfg.KafkaTopic, cfg.KafkaGroupID, statsUpdater)
	defer consumer.Close()

	expenseRepo := expensesrepo.NewPgxRepository(pool)
	expenseSvc := expensesservice.New(expenseRepo, producer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// start kafka consumer in background
	go consumer.Run(ctx)

	r := setupRouter(pool, expenseSvc)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Port)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()

	log.Printf("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
}
