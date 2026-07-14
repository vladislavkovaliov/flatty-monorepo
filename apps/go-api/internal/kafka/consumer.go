package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type StatsUpdater interface {
	UpsertTotal(ctx context.Context, month, year int, totalSpent float64) error
	UpsertAverage(ctx context.Context, month, year int, averageAmount float64, expenseCount int) error
	GetTotal(ctx context.Context, month, year int) (float64, error)
	GetAverage(ctx context.Context, month, year int) (float64, int, error)
}

type statsRepo struct {
	pool *pgxpool.Pool
}

func NewStatsRepoFromPool(pool *pgxpool.Pool) StatsUpdater {
	return &statsRepo{pool: pool}
}

func (r *statsRepo) GetTotal(ctx context.Context, month, year int) (float64, error) {
	var total float64
	err := r.pool.QueryRow(ctx, `SELECT total_spent FROM expense_monthly_totals WHERE month = $1 AND year = $2`, month, year).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (r *statsRepo) GetAverage(ctx context.Context, month, year int) (float64, int, error) {
	var avg float64
	var cnt int
	err := r.pool.QueryRow(ctx, `SELECT average_amount, expense_count FROM expense_monthly_averages WHERE month = $1 AND year = $2`, month, year).Scan(&avg, &cnt)
	if err != nil {
		return 0, 0, err
	}
	return avg, cnt, nil
}

func (r *statsRepo) UpsertTotal(ctx context.Context, month, year int, totalSpent float64) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_monthly_totals (month, year, total_spent, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (month, year) DO UPDATE SET
			total_spent = $3,
			updated_at = NOW()
	`, month, year, totalSpent)
	return err
}

func (r *statsRepo) UpsertAverage(ctx context.Context, month, year int, averageAmount float64, expenseCount int) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO expense_monthly_averages (month, year, average_amount, expense_count, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (month, year) DO UPDATE SET
			average_amount = $3,
			expense_count = $4,
			updated_at = NOW()
	`, month, year, averageAmount, expenseCount)
	return err
}

type Consumer struct {
	reader *kafka.Reader
	stats  StatsUpdater
}

func NewConsumer(brokers []string, topic, groupID string, stats StatsUpdater) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 10,
			MaxBytes: 10e6,
		}),
		stats: stats,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	log.Printf("kafka consumer started, topic: %s, group: %s", c.reader.Config().Topic, c.reader.Config().GroupID)

	for {
		select {
		case <-ctx.Done():
			log.Println("kafka consumer shutting down")
			return
		default:
		}

		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var event ExpenseEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("kafka unmarshal error: %v", err)
			continue
		}

		if err := c.processEvent(ctx, event); err != nil {
			log.Printf("kafka process error: %v", err)
		}
	}
}

func (c *Consumer) processEvent(ctx context.Context, event ExpenseEvent) error {
	switch event.Action {
	case "created":
		return c.handleCreated(ctx, event)
	case "updated":
		return c.handleUpdated(ctx, event)
	case "deleted":
		return c.handleDeleted(ctx, event)
	default:
		return fmt.Errorf("unknown action: %s", event.Action)
	}
}

func (c *Consumer) handleCreated(ctx context.Context, event ExpenseEvent) error {
	prevTotal, _ := c.stats.GetTotal(ctx, event.Month, event.Year)
	newTotal := prevTotal + event.Amount
	if err := c.stats.UpsertTotal(ctx, event.Month, event.Year, newTotal); err != nil {
		return fmt.Errorf("upsert total: %w", err)
	}

	prevAvg, prevCount, _ := c.stats.GetAverage(ctx, event.Month, event.Year)
	newCount := prevCount + 1
	newAvg := newTotal / float64(newCount)
	if prevCount == 0 {
		newAvg = event.Amount
	}
	if err := c.stats.UpsertAverage(ctx, event.Month, event.Year, newAvg, newCount); err != nil {
		return fmt.Errorf("upsert average: %w", err)
	}

	_ = prevAvg
	return nil
}

func (c *Consumer) handleUpdated(ctx context.Context, event ExpenseEvent) error {
	prevTotal, _ := c.stats.GetTotal(ctx, event.Month, event.Year)
	diff := event.Amount - event.PrevAmount
	newTotal := prevTotal + diff
	if err := c.stats.UpsertTotal(ctx, event.Month, event.Year, newTotal); err != nil {
		return fmt.Errorf("upsert total: %w", err)
	}

	_, prevCount, _ := c.stats.GetAverage(ctx, event.Month, event.Year)
	if prevCount > 0 {
		newAvg := newTotal / float64(prevCount)
		if err := c.stats.UpsertAverage(ctx, event.Month, event.Year, newAvg, prevCount); err != nil {
			return fmt.Errorf("upsert average: %w", err)
		}
	}

	return nil
}

func (c *Consumer) handleDeleted(ctx context.Context, event ExpenseEvent) error {
	prevTotal, _ := c.stats.GetTotal(ctx, event.Month, event.Year)
	newTotal := prevTotal - event.Amount
	if newTotal < 0 {
		newTotal = 0
	}
	if err := c.stats.UpsertTotal(ctx, event.Month, event.Year, newTotal); err != nil {
		return fmt.Errorf("upsert total: %w", err)
	}

	_, prevCount, _ := c.stats.GetAverage(ctx, event.Month, event.Year)
	newCount := prevCount - 1
	if newCount < 0 {
		newCount = 0
	}
	var newAvg float64
	if newCount > 0 {
		newAvg = newTotal / float64(newCount)
	}
	if err := c.stats.UpsertAverage(ctx, event.Month, event.Year, newAvg, newCount); err != nil {
		return fmt.Errorf("upsert average: %w", err)
	}

	return nil
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
