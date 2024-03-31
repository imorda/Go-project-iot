package domain

import "time"

type SensorType string

const (
	SensorTypeContactClosure SensorType = "cc"
	SensorTypeADC            SensorType = "adc"
)

// Sensor - структура для хранения данных датчика
type Sensor struct {
	ID           int64
	SerialNumber string
	Type         SensorType
	CurrentState int64
	Description  string
	IsActive     bool
	RegisteredAt time.Time
	LastActivity time.Time
}
