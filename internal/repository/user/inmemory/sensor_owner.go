package inmemory

import (
	"context"
	"homework/internal/domain"
	"sync"
)

type SensorOwnerRepository struct {
	storage []domain.SensorOwner
	mu      sync.Mutex
}

func NewSensorOwnerRepository() *SensorOwnerRepository {
	return &SensorOwnerRepository{
		storage: make([]domain.SensorOwner, 0),
		mu:      sync.Mutex{},
	}
}

func (r *SensorOwnerRepository) SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	r.storage = append(r.storage, sensorOwner)
	r.mu.Unlock()

	return nil
}

func (r *SensorOwnerRepository) GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([]domain.SensorOwner, 0)
	r.mu.Lock()
	for _, so := range r.storage {
		if so.UserID == userID {
			result = append(result, so)
		}
	}
	r.mu.Unlock()

	return result, nil
}
