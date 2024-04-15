package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
)

type EventRepository struct {
	storage map[int64]domain.Event
	mu      sync.Mutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		storage: make(map[int64]domain.Event),
		mu:      sync.Mutex{},
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event == nil {
		return errors.New("got nil event at SaveEvent()")
	}

	r.mu.Lock()
	if cur, exists := r.storage[event.SensorID]; !exists || cur.Timestamp.Before(event.Timestamp) {
		r.storage[event.SensorID] = *event
	}
	r.mu.Unlock()

	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	val, exists := r.storage[id]
	r.mu.Unlock()
	if !exists {
		return nil, usecase.ErrEventNotFound
	}

	return &val, nil
}
