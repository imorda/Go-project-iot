package usecase

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
)

type Event struct {
	eventRepository  EventRepository
	sensorRepository SensorRepository
}

func NewEvent(er EventRepository, sr SensorRepository) *Event {
	return &Event{
		eventRepository:  er,
		sensorRepository: sr,
	}
}

func (e *Event) ReceiveEvent(ctx context.Context, event *domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event == nil {
		return errors.New("got nil event at ReceiveEvent()")
	}
	if event.Timestamp.IsZero() {
		return ErrInvalidEventTimestamp
	}
	sensor, err := e.sensorRepository.GetSensorBySerialNumber(ctx, event.SensorSerialNumber)
	if err != nil {
		return fmt.Errorf("invalid sensor serial number in event %v: %w", event, err)
	}
	event.SensorID = sensor.ID

	if err := e.eventRepository.SaveEvent(ctx, event); err != nil {
		return fmt.Errorf("cannot save event %v: %w", event, err)
	}

	sensor.CurrentState = event.Payload
	sensor.LastActivity = event.Timestamp
	if err := e.sensorRepository.SaveSensor(ctx, sensor); err != nil {
		return fmt.Errorf("cannot save new sensor state %v: %w", sensor, err)
	}

	return nil
}

func (e *Event) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	event, err := e.eventRepository.GetLastEventBySensorID(ctx, id)
	if err != nil {
		return event, fmt.Errorf("cannot get last event for id %v: %w", id, err)
	}

	return event, err
}
