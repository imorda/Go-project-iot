package http

import (
	"context"
	"encoding/json"
	"homework/internal/domain"
	"homework/internal/usecase"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nhooyr.io/websocket"
)

func TestServerConfiguration(t *testing.T) {
	uc := UseCases{}
	s := NewServer(uc, WithHost("localhost"), WithPort(8765))

	assert.Equal(t, "localhost", s.host)
	assert.Equal(t, uint16(8765), s.port)
	assert.NotNil(t, s.wsh)
	assert.NotNil(t, s.router)
}

func TestConnectionAndSuccessfulFinalization(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	erMock := usecase.NewMockEventRepository(ctrl)
	erMock.EXPECT().GetLastEventBySensorID(gomock.Any(), gomock.Eq(int64(1))).Return(&domain.Event{SensorID: 1, Payload: 100}, nil).Times(1)
	srMock := usecase.NewMockSensorRepository(ctrl)
	srMock.EXPECT().GetSensorByID(gomock.Any(), gomock.Eq(int64(1))).Return(&domain.Sensor{ID: 1}, nil).Times(1)
	urMock := usecase.NewMockUserRepository(ctrl)
	sorMock := usecase.NewMockSensorOwnerRepository(ctrl)

	esrMock := usecase.NewMockSubscriptionRepository[domain.Event](ctrl)
	dummyChan := make(chan domain.Event, 1)
	defer close(dummyChan)
	subscriptionId := uuid.New()
	esrMock.EXPECT().Subscribe(gomock.Any(), gomock.Eq(int64(1))).Return(&domain.Subscription[domain.Event]{
		SubscriptionWriteHandle: domain.SubscriptionWriteHandle[domain.Event]{
			Ch: dummyChan,
		},
		SubscriptionReadHandle: domain.SubscriptionReadHandle[domain.Event]{
			Ch: dummyChan,
		},
		SensorID: 1,
		Id:       subscriptionId,
	}, nil).Times(1)
	esrMock.EXPECT().Unsubscribe(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(subscriptionId)).Return(nil).Times(1)

	uc := UseCases{
		Event:             usecase.NewEvent(erMock, srMock, esrMock),
		Sensor:            usecase.NewSensor(srMock),
		User:              usecase.NewUser(urMock, sorMock, srMock),
		EventSubscription: usecase.NewSubscription[domain.Event](esrMock, srMock),
	}

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.NoError(t, NewServer(uc, WithHost("localhost"), WithPort(8765)).Run(ctx, cancel))
	}()

	defer wg.Wait()
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws://localhost:8765/api/sensors/1/events", nil)
	require.NoError(t, err)
	op, msg, err := conn.Read(ctx)
	require.NoError(t, err)
	require.Equal(t, websocket.MessageText, op)
	var event domain.Event
	require.NoError(t, json.Unmarshal(msg, &event))
	require.Equal(t, int64(1), event.SensorID)
	require.Equal(t, int64(100), event.Payload)
	conn.CloseRead(ctx)
}
