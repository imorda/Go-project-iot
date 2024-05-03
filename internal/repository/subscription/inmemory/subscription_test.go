package inmemory

import (
	"context"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestSubscriptionRepository_SaveEvent(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := sr.Subscribe(ctx, 1)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := sr.Subscribe(ctx, 1)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("ok", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ret, err := sr.Subscribe(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, ret)

		event := domain.Event{
			Timestamp: time.Now(),
			SensorID:  228,
			Payload:   0,
		}

		assert.Equal(t, int64(1), ret.SensorID)
		ret.SubscriptionWriteHandle.Ch <- event
		actualEvent := <-ret.SubscriptionReadHandle.Ch

		assert.Equal(t, event, actualEvent)
	})

	t.Run("ok, collision test", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		wg := sync.WaitGroup{}
		var sut *domain.Subscription[domain.Event]
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			curI := i
			go func() {
				defer wg.Done()
				ret, err := sr.Subscribe(ctx, int64(curI))
				assert.NoError(t, err)
				assert.NotNil(t, ret)
				if curI == 228 {
					sut = ret
				}
			}()
		}

		wg.Wait()

		broadcast, err := sr.GetBroadcastHandleById(ctx, 228)
		assert.NoError(t, err)
		assert.NotNil(t, broadcast)
		assert.Equal(t, int64(228), sut.SensorID)

		event := domain.Event{
			Timestamp: time.Now(),
			SensorID:  228,
			Payload:   0,
		}
		broadcast.Ch <- event
		actualEvent := <-sut.SubscriptionReadHandle.Ch

		assert.Equal(t, event, actualEvent)
	})
}

func TestSubscriptionRepository_GetBroadcastHandleById(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := sr.GetBroadcastHandleById(ctx, 0)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		_, err := sr.GetBroadcastHandleById(ctx, 0)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, sensor not found", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		_, err := sr.GetBroadcastHandleById(context.Background(), 234)
		assert.ErrorIs(t, err, usecase.ErrSensorNotFound)
	})

	t.Run("ok, broadcast to multiple subscribers", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		goodHandles := make([]domain.Subscription[domain.Event], 0)
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			ret, err := sr.Subscribe(ctx, 12345)
			assert.NoError(t, err)
			assert.NotNil(t, ret)
			goodHandles = append(goodHandles, *ret)
		}

		badHandles := make([]domain.Subscription[domain.Event], 0)
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			ret, err := sr.Subscribe(ctx, 54321)
			assert.NoError(t, err)
			assert.NotNil(t, ret)
			badHandles = append(badHandles, *ret)
		}

		broadcast, err := sr.GetBroadcastHandleById(ctx, 12345)
		assert.NoError(t, err)
		assert.NotNil(t, broadcast)

		event := domain.Event{
			Timestamp: time.Now(),
			SensorID:  12345,
			Payload:   0,
		}
		broadcast.Ch <- event

		wg := sync.WaitGroup{}

		for _, i := range goodHandles {
			wg.Add(1)
			curSub := i
			go func() {
				defer wg.Done()

				actualEvent := <-curSub.SubscriptionReadHandle.Ch
				assert.Equal(t, event, actualEvent)
			}()
		}

		wg.Wait()

		for _, i := range badHandles {
			wg.Add(1)
			curSub := i
			go func() {
				defer wg.Done()

				select {
				case x, ok := <-curSub.SubscriptionReadHandle.Ch:
					if ok {
						assert.Failf(t, "Unexpected event", "Read event for sensor %d: %v", x.SensorID, x.Payload)
					} else {
						assert.Fail(t, "Channel closed")
					}
				default:
				}
			}()
		}
		wg.Wait()
	})
}

func TestSubscriptionRepository_Unsubscribe(t *testing.T) {
	t.Run("fail, ctx cancelled", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := sr.Unsubscribe(ctx, 0, uuid.New())
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("fail, ctx deadline exceeded", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		err := sr.Unsubscribe(ctx, 0, uuid.New())
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("fail, sensor not found", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		err := sr.Unsubscribe(context.Background(), 0, uuid.New())
		assert.ErrorIs(t, err, usecase.ErrSensorNotFound)
	})

	t.Run("fail, not subscribed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewSubscriptionRepository[domain.Event]()
		ret, err := sr.Subscribe(ctx, 0)
		assert.NoError(t, err)
		assert.NotNil(t, ret)
		id := ret.Id

		err = sr.Unsubscribe(ctx, 0, id)
		assert.NoError(t, err)
		err = sr.Unsubscribe(ctx, 0, id)
		assert.ErrorIs(t, err, usecase.ErrSubscriptionNotFound)
	})

	t.Run("ok, broadcast to multiple subscribers", func(t *testing.T) {
		sr := NewSubscriptionRepository[domain.Event]()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		goodHandles := make([]domain.Subscription[domain.Event], 0)
		var badHandle domain.Subscription[domain.Event]
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			ret, err := sr.Subscribe(ctx, 12345)
			assert.NoError(t, err)
			assert.NotNil(t, ret)
			if i == 7 {
				badHandle = *ret
			} else {
				goodHandles = append(goodHandles, *ret)
			}
		}

		err := sr.Unsubscribe(ctx, badHandle.SensorID, badHandle.Id)
		assert.NoError(t, err)

		broadcast, err := sr.GetBroadcastHandleById(ctx, 12345)
		assert.NoError(t, err)
		assert.NotNil(t, broadcast)

		event := domain.Event{
			Timestamp: time.Now(),
			SensorID:  12345,
			Payload:   0,
		}
		broadcast.Ch <- event

		wg := sync.WaitGroup{}

		for _, i := range goodHandles {
			wg.Add(1)
			curSub := i
			go func() {
				defer wg.Done()

				actualEvent := <-curSub.SubscriptionReadHandle.Ch
				assert.Equal(t, event, actualEvent)
			}()
		}

		wg.Wait()

		select {
		case x, ok := <-badHandle.SubscriptionReadHandle.Ch:
			if ok {
				assert.Failf(t, "Unexpected event", "Read event for sensor %d: %v", x.SensorID, x.Payload)
			}
		default:
			assert.Fail(t, "Channel not closed")
		}
	})
}
