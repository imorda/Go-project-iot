package domain

// User - структура для хранения пользователя
type User struct {
	ID   int64
	Name string
}

// SensorOwner - структура для связи пользователя и датчика
// UserID - id пользователя
// SensorID - id датчика
// Связь многие-ко-многим: пользователь может иметь доступ к нескольким датчикам, датчик может быть доступен для нескольких пользователей.
type SensorOwner struct {
	UserID   int64
	SensorID int64
}
