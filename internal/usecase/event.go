package usecase

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"time"
)

type Event struct {
	eventRepository             EventRepository
	sensorRepository            SensorRepository
	eventSubscriptionRepository SubscriptionRepository[domain.Event]
}

func NewEvent(er EventRepository, sr SensorRepository, esr SubscriptionRepository[domain.Event]) *Event {
	return &Event{
		eventRepository:             er,
		sensorRepository:            sr,
		eventSubscriptionRepository: esr,
	}
}

func (e *Event) broadcastEvent(ctx context.Context, sensId int64, event *domain.Event) error {
	handle, err := e.eventSubscriptionRepository.GetBroadcastHandleById(ctx, sensId)
	if err != nil {
		if errors.Is(err, ErrSensorNotFound) {
			return nil
		}
		return fmt.Errorf("can't get event broadcast handle: %w", err)
	}

	handle.Ch <- *event
	return nil
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if event == nil {
		return errors.New("got nil event at ReceiveEvent()")
	}
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}
	sens, err := e.sensorRepository.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return fmt.Errorf("invalid sensor serial number in event %v: %w", event, err)
	}
	event.SensorID = sens.ID

	if err := e.eventRepository.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("cannot save event %v: %w", event, err)
	}

	sens.CurrentState = event.Payload
	sens.LastActivity = event.Timestamp
	if err := e.sensorRepository.SaveSensor(ctx, sens); err != nil {
		return fmt.Errorf("cannot save new sensor state %v: %w", sens, err)
	}

	return e.broadcastEvent(ctx, sens.ID, event)
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	event, err := e.eventRepository.GetLastEventBySensorID(ctx, id)
	if err != nil {
		return event, fmt.Errorf("cannot get last event for id %v: %w", id, err)
	}

	return event, err
}

func (e *Event) GetEventsHistoryBySensorID(ctx context.Context, id int64, startTime, endTime time.Time) ([]*domain.Event, error) {
	if startTime.After(endTime) {
		return nil, ErrInvalidEventTimestamp
	}
	events, err := e.eventRepository.GetEventsHistoryBySensorID(ctx, id, startTime, endTime)
	if err != nil {
		return events, fmt.Errorf("cannot get events history for id %v: %w", id, err)
	}

	return events, err
}
