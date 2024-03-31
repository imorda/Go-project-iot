package domain

import "time"

// Event - структура события по датчику
type Event struct {
	Timestamp          time.Time
	SensorSerialNumber string
	SensorID           int64
	Payload            int64
}
