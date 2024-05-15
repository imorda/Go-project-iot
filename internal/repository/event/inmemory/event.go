package inmemory

import (
	"context"
	"errors"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"time"

	"github.com/hashicorp/go-set/v2"
)

type EventRepository struct {
	storage map[int64]*set.TreeSet[*domain.Event]
	mu      sync.Mutex
}

func compareEvent(lhs, rhs *domain.Event) int {
	cmp1 := lhs.Timestamp.Compare(rhs.Timestamp)
	if cmp1 == 0 {
		return set.Compare(lhs.Payload, rhs.Payload)
	}
	return cmp1
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		storage: make(map[int64]*set.TreeSet[*domain.Event]),
		mu:      sync.Mutex{},
	}
}

func (r *EventRepository) SaveEvent(ctx context.Context, event *domain.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if event == nil {
		return errors.New("got nil event at SaveEvent()")
	}

	r.mu.Lock()
	events, exists := r.storage[event.SensorID]
	if !exists {
		events = set.NewTreeSet[*domain.Event](compareEvent)
		r.storage[event.SensorID] = events
	}
	events.Insert(event)
	r.mu.Unlock()

	return nil
}

func (r *EventRepository) GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	val, exists := r.storage[id]
	if !exists || val.Empty() {
		return nil, usecase.ErrEventNotFound
	}

	return val.Max(), nil
}

func (r *EventRepository) GetEventsHistoryBySensorID(ctx context.Context, id int64, startTime, endTime time.Time) ([]*domain.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	val, exists := r.storage[id]
	if !exists || val.Empty() {
		return nil, nil // Empty result if no events ever existed for this sensor, no error
	}

	return val.AboveEqual(&domain.Event{Timestamp: startTime}).Below(&domain.Event{Timestamp: endTime.Add(time.Nanosecond)}).Slice(), nil
}
