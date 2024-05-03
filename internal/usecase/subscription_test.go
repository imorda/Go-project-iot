package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_subscription_Subscribe(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("err, sensor not found", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorByID(ctx, gomock.Any()).Times(1).Return(nil, ErrSensorNotFound)

		s := NewSubscription[domain.Event](nil, sr)

		_, err := s.Subscribe(ctx, 1)
		assert.ErrorIs(t, err, ErrSensorNotFound)
	})

	t.Run("err, subscribe error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorByID(ctx, gomock.Eq(int64(3))).Times(1).Return(&domain.Sensor{
			ID:           3,
			SerialNumber: "123",
		}, nil)

		esr := NewMockSubscriptionRepository[domain.Event](ctrl)
		expectedError := errors.New("some error")
		esr.EXPECT().Subscribe(ctx, gomock.Eq(int64(3))).Times(1).Return(nil, expectedError)

		s := NewSubscription[domain.Event](esr, sr)

		_, err := s.Subscribe(ctx, 3)
		assert.ErrorIs(t, err, expectedError)
	})

	t.Run("ok, no error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sr := NewMockSensorRepository(ctrl)

		sr.EXPECT().GetSensorByID(ctx, gomock.Eq(int64(3))).Times(1).Return(&domain.Sensor{
			ID:           3,
			SerialNumber: "123",
		}, nil)

		expectedSubscription := domain.Subscription[domain.Event]{
			Id: uuid.New(),
		}
		esr := NewMockSubscriptionRepository[domain.Event](ctrl)
		esr.EXPECT().Subscribe(ctx, gomock.Eq(int64(3))).Times(1).Return(&expectedSubscription, nil)

		s := NewSubscription[domain.Event](esr, sr)
		sub, err := s.Subscribe(ctx, 3)
		assert.NoError(t, err)
		assert.Equal(t, expectedSubscription, *sub)
	})
}
