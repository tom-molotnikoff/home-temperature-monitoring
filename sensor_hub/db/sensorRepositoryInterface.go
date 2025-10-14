package database

type SensorRepositoryInterface[T any] interface {
	AddSensor(sensor T) error
	UpdateSensorByName(sensor T) error
	DeleteSensorByName(name string) error
	GetSensorByName(name string) (*T, error)
	GetAllSensors() ([]T, error)
	GetSensorsByType(sensorType string) ([]T, error)
	GetSensorIdByName(name string) (int, error)
	SensorExists(name string) (bool, error)
}
