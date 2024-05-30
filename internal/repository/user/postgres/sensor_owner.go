package postgres

import (
	"context"
	"fmt"
	"homework/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SensorOwnerRepository struct {
	pool *pgxpool.Pool
}

func NewSensorOwnerRepository(pool *pgxpool.Pool) *SensorOwnerRepository {
	return &SensorOwnerRepository{
		pool,
	}
}

const saveSensorOwnerQuery = `INSERT INTO sensors_users (sensor_id, user_id) VALUES ($1, $2)`

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	if _, err := r.pool.Exec(ctx, saveSensorOwnerQuery, sensorOwner.SensorID, sensorOwner.UserID); err != nil {
		return fmt.Errorf("unable to save sensor owner to pg: %w", err)
	}
	return nil
}

const getSensorsByUserIDQuery = `SELECT sensor_id, user_id FROM sensors_users WHERE user_id = $1`

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	rows, err := r.pool.Query(ctx, getSensorsByUserIDQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("can't get sensors: %w", err)
	}

	defer rows.Close()

	var result []domain.SensorOwner
	for rows.Next() {
		sensor := domain.SensorOwner{}
		if err := rows.Scan(&sensor.SensorID, &sensor.UserID); err != nil {
			return nil, fmt.Errorf("can't scan sensors: %w", err)
		}

		result = append(result, sensor)
	}

	return result, nil
}
