package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
)

var ErrSensorNotFound = errors.New("sensor not found")

type SensorRepository struct {
	// TODO добавьте реализацию
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{}
}

func (r *SensorRepository) SaveSensor(ctx context.Context, sensor *domain.Sensor) error {
	// TODO добавьте реализацию
	return nil
}

func (r *SensorRepository) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, nil
}

func (r *SensorRepository) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, nil
}

func (r *SensorRepository) GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, ErrSensorNotFound
}
