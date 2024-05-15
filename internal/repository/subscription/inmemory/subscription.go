package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"

	"github.com/google/uuid"
)

type SubscriptionRepository[T any] struct {
	storage map[int64]Subscribers[T]
	mu      sync.Mutex
}

type Subscribers[T any] struct {
	SubscribersMap map[uuid.UUID]domain.Subscription[T]
}

func NewSubscriptionRepository[T any]() *SubscriptionRepository[T] {
	return &SubscriptionRepository[T]{
		storage: make(map[int64]Subscribers[T]),
		mu:      sync.Mutex{},
	}
}

func (sr *SubscriptionRepository[T]) Subscribe(ctx context.Context, subscription domain.Subscription[T]) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	sm, ok := sr.storage[subscription.SensorID]
	if !ok {
		sm = Subscribers[T]{
			SubscribersMap: map[uuid.UUID]domain.Subscription[T]{},
		}
		sr.storage[subscription.SensorID] = sm
	}

	sm.SubscribersMap[subscription.Id] = subscription
	return nil
}

func (sr *SubscriptionRepository[T]) Unsubscribe(ctx context.Context, sensId int64, subscriptionId uuid.UUID) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	sm, ok := sr.storage[sensId]
	if !ok {
		return usecase.ErrSensorNotFound
	}

	delete(sm.SubscribersMap, subscriptionId)

	return nil
}

func (sr *SubscriptionRepository[T]) GetBroadcastHandleById(ctx context.Context, id int64) (*domain.SubscriptionWriteHandle[T], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sr.mu.Lock()
	defer sr.mu.Unlock()

	in := make(chan T)

	sm, ok := sr.storage[id]
	if !ok {
		return nil, usecase.ErrSensorNotFound
	}

	go func() {
		for upd := range in {
			sr.mu.Lock() // Makes sure nobody unsubscribes during the broadcast, all writes need to be nonblocking
			for _, s := range sm.SubscribersMap {
				s.SubscriptionWriteHandle.Ch <- upd
			}
			sr.mu.Unlock()
		}
	}()

	return &domain.SubscriptionWriteHandle[T]{
		Ch: in,
	}, nil
}
