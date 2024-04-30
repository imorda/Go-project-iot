package domain

import (
	"github.com/google/uuid"
)

type Subscription[T any] struct {
	SubscriptionWriteHandle[T]
	SubscriptionReadHandle[T]
	SensorID int64
	Id       uuid.UUID
}

type SubscriptionReadHandle[T any] struct {
	Ch <-chan T
}

type SubscriptionWriteHandle[T any] struct {
	Ch chan<- T
}
