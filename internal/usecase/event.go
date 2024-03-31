package usecase

import (
	"context"
	"homework/internal/domain"
)

type Event struct {
	// TODO добавьте реализацию
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	// TODO добавьте реализацию
	return nil
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	// TODO добавьте реализацию
	return nil, nil
}
