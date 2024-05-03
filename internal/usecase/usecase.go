package usecase

import (
	"context"
	"errors"
	"homework/internal/domain"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWrongSensorSerialNumber = errors.New("wrong sensor serial number")
	ErrWrongSensorType         = errors.New("wrong sensor type")
	ErrInvalidEventTimestamp   = errors.New("invalid event timestamp")
	ErrInvalidUserName         = errors.New("invalid user name")
	ErrSensorNotFound          = errors.New("sensor not found")
	ErrUserNotFound            = errors.New("user not found")
	ErrEventNotFound           = errors.New("event not found")
	ErrSubscriptionNotFound    = errors.New("subscription not found")
)

// requires mockgen v1.7+

//go:generate mockgen -source usecase.go -package usecase -destination usecase_mock.go
type SensorRepository interface {
	// SaveSensor - функция сохранения датчика
	SaveSensor(ctx context.Context, sensor *domain.Sensor) error
	// GetSensors - функция получения списка датчиков
	GetSensors(ctx context.Context) ([]domain.Sensor, error)
	// GetSensorByID - функция получения датчика по ID
	GetSensorByID(ctx context.Context, id int64) (*domain.Sensor, error)
	// GetSensorBySerialNumber - функция получения датчика по серийному номеру
	GetSensorBySerialNumber(ctx context.Context, sn string) (*domain.Sensor, error)
}

type SubscriptionRepository[T any] interface {
	// Subscribe - функция подписки на какое-то изменение датчика
	Subscribe(ctx context.Context, sensorId int64) (*domain.Subscription[T], error)
	// Unsubscribe - функция отмены подписки
	Unsubscribe(ctx context.Context, sensId int64, subscriptionId uuid.UUID) error
	// GetBroadcastHandleById - функция получения "ручки" для оповещения всех подписчиков об изменении
	GetBroadcastHandleById(ctx context.Context, id int64) (*domain.SubscriptionWriteHandle[T], error)
}

type EventRepository interface {
	// SaveEvent - функция сохранения события по датчику
	SaveEvent(ctx context.Context, event *domain.Event) error
	// GetLastEventBySensorID - функция получения последнего события по ID датчика
	GetLastEventBySensorID(ctx context.Context, id int64) (*domain.Event, error)
	// GetEventsHistoryBySensorID - функция получения истории событий по ID датчика и временному промежутку
	GetEventsHistoryBySensorID(ctx context.Context, id int64, startTime, endTime time.Time) ([]*domain.Event, error)
}

type UserRepository interface {
	// SaveUser - функция сохранения пользователя
	SaveUser(ctx context.Context, user *domain.User) error
	// GetUserByID - функция получения пользователя по id
	GetUserByID(ctx context.Context, id int64) (*domain.User, error)
}

type SensorOwnerRepository interface {
	// SaveSensorOwner - функция привязки датчика к пользователю
	SaveSensorOwner(ctx context.Context, sensorOwner domain.SensorOwner) error
	// GetSensorsByUserID -функция, возвращающая список привязок для пользователя
	GetSensorsByUserID(ctx context.Context, userID int64) ([]domain.SensorOwner, error)
}
