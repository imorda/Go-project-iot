package usecase

import (
	"context"
	"homework/internal/domain"

	"github.com/google/uuid"
)

type Subscription[T any] struct {
	subscriptionRepository SubscriptionRepository[T]
	sensorRepository       SensorRepository
}

func NewSubscription[T any](sur SubscriptionRepository[T], ser SensorRepository) *Subscription[T] {
	return &Subscription[T]{
		subscriptionRepository: sur,
		sensorRepository:       ser,
	}
}

func (s *Subscription[T]) Subscribe(ctx context.Context, sensId int64) (*domain.Subscription[T], error) {
	if _, err := s.sensorRepository.GetSensorByID(ctx, sensId); err != nil {
		return nil, err
	}

	return s.subscriptionRepository.Subscribe(ctx, sensId)
}

func (s *Subscription[T]) Unsubscribe(ctx context.Context, sensId int64, subscriptionId uuid.UUID) error {
	return s.subscriptionRepository.Unsubscribe(ctx, sensId, subscriptionId)
}
