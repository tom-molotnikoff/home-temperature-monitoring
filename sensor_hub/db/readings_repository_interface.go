package database

import "time"

type ReadingsRepository[T any] interface {
	Add(readings []T) error
	GetBetweenDates(tableName, startDate, endDate string) ([]T, error)
	GetLatest() ([]T, error)
	GetTotalReadingsBySensorId(sensorId int) (int, error)
	DeleteReadingsOlderThan(cutoffDate time.Time) error
}
