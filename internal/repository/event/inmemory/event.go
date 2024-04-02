package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"sync"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	storage map[int64][]domain.Event
	mu      sync.Mutex
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		storage: make(map[int64][]domain.Event),
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
	slice, exists := r.storage[event.SensorID]
	if !exists {
		slice = make([]domain.Event, 0, 1)
	}

	r.storage[event.SensorID] = append(slice, *event)
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
		return nil, ErrEventNotFound
	}

	return &val[len(val)-1], nil
}
