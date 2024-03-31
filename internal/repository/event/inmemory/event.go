package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
)

var ErrEventNotFound = errors.New("event not found")

type EventRepository struct {
	// TODO добавьте реализацию
}

func NewEventRepository() *EventRepository {
	return &EventRepository{}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	// TODO добавьте реализацию
	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	// TODO добавьте реализацию
	return nil, nil
}
