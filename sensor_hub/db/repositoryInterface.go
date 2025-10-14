package database

type Repository[T any] interface {
	Add(readings []T) error
	GetBetweenDates(tableName, startDate, endDate string) ([]T, error)
	GetLatest() ([]T, error)
}
