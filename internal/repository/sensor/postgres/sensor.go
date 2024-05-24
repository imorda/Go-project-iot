package postgres

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/domain"
	"homework/internal/usecase"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorRepository struct {
	pool *pgxpool.Pool
}

func NewSensorRepository(pool *pgxpool.Pool) *SensorRepository {
	return &SensorRepository{
		pool: pool,
	}
}

const saveSensorQuery = `
	INSERT INTO sensors (serial_number, type, current_state, description, is_active, registered_at, last_activity) 
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (serial_number) DO UPDATE 
	  SET serial_number = excluded.serial_number, 
		  type = excluded.type,
		  current_state = excluded.current_state,
		  description = excluded.description,
		  is_active = excluded.is_active,
		  last_activity = excluded.last_activity
	RETURNING id`

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	row := r.pool.QueryRow(ctx, saveSensorQuery, sensor.SerialNumber, sensor.Type, sensor.CurrentState,
		sensor.Description, sensor.IsActive, time.Now(), sensor.LastActivity)

	if err := row.Scan(&sensor.ID); err != nil {
		return fmt.Errorf("unable to save sensor to pg: %w", err)
	}

	return nil
}

const getSensorsQuery = `
	SELECT 
	    id, serial_number, type, current_state, description, is_active, registered_at, last_activity 
	FROM sensors`

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	rows, err := r.pool.Query(ctx, getSensorsQuery)
	if err != nil {
		return nil, fmt.Errorf("can't get sensors: %w", err)
	}

	defer rows.Close()

	var result []domain.Sensor
	for rows.Next() {
		sensor := domain.Sensor{}
		if err := rows.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description,
			&sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity); err != nil {
			return nil, fmt.Errorf("can't scan sensors: %w", err)
		}

		result = append(result, sensor)
	}

	return result, nil
}

const getSensorByIDQuery = `
	SELECT 
	    id, serial_number, type, current_state, description, is_active, registered_at, last_activity 
	FROM sensors 
	WHERE id = $1`

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, getSensorByIDQuery, id)

	sensor := &domain.Sensor{}
	if err := row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description,
		&sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, fmt.Errorf("unable to find sensor by id: %w", err)
	}

	return sensor, nil
}

const getSensorBySerialNumberQuery = `
	SELECT 
	    id, serial_number, type, current_state, description, is_active, registered_at, last_activity 
	FROM sensors 
	WHERE serial_number = $1`

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	row := r.pool.QueryRow(ctx, getSensorBySerialNumberQuery, sn)

	sensor := &domain.Sensor{}
	if err := row.Scan(&sensor.ID, &sensor.SerialNumber, &sensor.Type, &sensor.CurrentState, &sensor.Description,
		&sensor.IsActive, &sensor.RegisteredAt, &sensor.LastActivity); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, usecase.ErrSensorNotFound
		}
		return nil, fmt.Errorf("unable to find sensor by serial number: %w", err)
	}

	return sensor, nil
}
