package usecase

import (
	"context"
	"fmt"
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

func (s *Subscription[T]) Subscribe(ctx context.Context, sensorId int64) (*domain.Subscription[T], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, err := s.sensorRepository.GetSensorByID(ctx, sensorId); err != nil {
		return nil, err
	}

	subscriptionId := uuid.New()
	ch := make(chan T, 1)

	subscription := domain.Subscription[T]{
		SubscriptionWriteHandle: domain.SubscriptionWriteHandle[T]{
			Ch: ch,
		},
		SubscriptionReadHandle: domain.SubscriptionReadHandle[T]{
			Ch: ch,
		},
		SensorID: sensorId,
		Id:       subscriptionId,
	}
	err := s.subscriptionRepository.Subscribe(ctx, subscription)
	if err != nil {
		close(ch)
		return nil, err
	}
	return &subscription, nil
}

func (s *Subscription[T]) Unsubscribe(ctx context.Context, sensId int64, subscriptionId uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	err := s.subscriptionRepository.Unsubscribe(ctx, sensId, subscriptionId)
	if err != nil {
		return fmt.Errorf("unable to unsubscribe %v from %v: %w", sensId, subscriptionId, err)
	}

	return nil
}
