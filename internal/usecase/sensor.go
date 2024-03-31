package usecase

import (
	"context"
	"homework/internal/domain"
)

type Sensor struct {
	// TODO добавьте реализацию
}

func NewSensor(sr SensorRepository) *Sensor {
	return &Sensor{}
}

func (s *Sensor) RegisterSensor(ctx context.Context, sensor *domain.Sensor) (*domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, nil
}

func (s *Sensor) GetSensors(ctx context.Context) ([]domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, nil
}

func (s *Sensor) GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error) {
	// TODO добавьте реализацию
	return nil, nil
}
