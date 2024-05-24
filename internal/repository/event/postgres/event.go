package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
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

const saveEventQuery = `INSERT INTO events (timestamp, sensor_serial_number, sensor_id, payload) VALUES ($1, $2, $3, $4)`

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if _, err := r.pool.Exec(ctx, saveEventQuery, event.Timestamp, event.SensorSerialNumber, event.SensorID, event.Payload); err != nil {
		return fmt.Errorf("unable to save event to pg: %w", err)
	}
	return nil
}

const getLastEventBySensorIDQuery = `
	SELECT timestamp, sensor_serial_number, sensor_id, payload FROM events WHERE sensor_id = $1 ORDER BY timestamp DESC`

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	row := r.pool.QueryRow(ctx, getLastEventBySensorIDQuery, id)

	event := &domain.Event{}
	if err := row.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEventNotFound
		}
		return nil, fmt.Errorf("unable to find event by sensor id: %w", err)
	}

	return event, nil
}

const getEventsHistoryBySensorIDQuery = `
	SELECT 
	    timestamp, sensor_serial_number, sensor_id, payload 
	FROM events 
	WHERE TRUE
	 AND sensor_id = $1
	 AND (timestamp BETWEEN $2 AND $3)`

func (r *EventRepository) GetEventsHistoryBySensorID(ctx context.Context, id int64, startTime, endTime time.Time) ([]*domain.Event, error) {
	rows, err := r.pool.Query(ctx, getEventsHistoryBySensorIDQuery, id, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("can't get events: %w", err)
	}

	defer rows.Close()

	var result []*domain.Event
	for rows.Next() {
		event := &domain.Event{}
		if err := rows.Scan(&event.Timestamp, &event.SensorSerialNumber, &event.SensorID, &event.Payload); err != nil {
			return nil, fmt.Errorf("can't scan sensors: %w", err)
		}

		result = append(result, event)
	}

	return result, nil
}
