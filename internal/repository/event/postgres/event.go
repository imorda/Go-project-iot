package postgres

import (
	"context"
	"errors"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{
		pool,
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	// TODO добавьте реализацию
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	// TODO добавьте реализацию
	return nil, nil
}
