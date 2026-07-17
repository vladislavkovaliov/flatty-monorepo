package expenses

import (
	"context"
	"fmt"

	kafkaclient "flatty-budget/go-api/internal/kafka"

	expensedomain "flatty-budget/go-api/domains/expenses"
)

type Service struct {
	repo     expensedomain.Repository
	producer kafkaclient.ProducerInterface
}

func New(repo expensedomain.Repository, producer kafkaclient.ProducerInterface) *Service {
	return &Service{
		repo:     repo,
		producer: producer,
	}
}

func (s *Service) Count(ctx context.Context) (int, error) {
	count, err := s.repo.Count(ctx)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*expensedomain.Expense, int, error) {
	items, err := s.repo.List(ctx, limit, offset)

	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	fmt.Println(total)

	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *Service) Create(ctx context.Context, input *expensedomain.ExpenseInput) (*expensedomain.Expense, error) {
	expense, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if s.producer != nil {
		event := kafkaclient.ExpenseEvent{
			Action: "created",
			ID:     expense.ID(),
			Month:  expense.Month(),
			Year:   expense.Year(),
			Amount: expense.Amount(),
		}
		if pubErr := s.producer.PublishEvent(ctx, event); pubErr != nil {
			fmt.Printf("failed to publish created event: %v\n", pubErr)
		}
	}

	return expense, nil
}

func (s *Service) Update(ctx context.Context, id int64, input *expensedomain.ExpenseInput) (*expensedomain.Expense, error) {
	prevExpense, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	expense, err := s.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if s.producer != nil {
		event := kafkaclient.ExpenseEvent{
			Action:     "updated",
			ID:         expense.ID(),
			Month:      expense.Month(),
			Year:       expense.Year(),
			Amount:     expense.Amount(),
			PrevAmount: prevExpense.Amount(),
		}
		if pubErr := s.producer.PublishEvent(ctx, event); pubErr != nil {
			fmt.Printf("failed to publish updated event: %v\n", pubErr)
		}
	}

	return expense, nil
}

func (s *Service) Delete(ctx context.Context, id int64) (int64, error) {
	prevExpense, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return 0, err
	}

	returningID, err := s.repo.Delete(ctx, id)
	if err != nil {
		return 0, err
	}

	if s.producer != nil {
		event := kafkaclient.ExpenseEvent{
			Action: "deleted",
			ID:     prevExpense.ID(),
			Month:  prevExpense.Month(),
			Year:   prevExpense.Year(),
			Amount: prevExpense.Amount(),
		}
		if pubErr := s.producer.PublishEvent(ctx, event); pubErr != nil {
			fmt.Printf("failed to publish deleted event: %v\n", pubErr)
		}
	}

	return returningID, nil
}
