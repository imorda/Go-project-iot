package inmemory

import (
	"context"
	"homework/internal/domain"
)

type SensorOwnerRepository struct {
	// TODO добавьте реализацию
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	// TODO добавьте реализацию
	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	// TODO добавьте реализацию
	return nil, nil
}
