package database

import (
	"errors"
	"testing"
	"time"

	"example/sensorHub/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSensorRepoForTemp implements SensorRepositoryInterface for testing TemperatureRepository
type MockSensorRepoForTemp struct {
	mock.Mock
}

func (m *MockSensorRepoForTemp) AddSensor(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorRepoForTemp) UpdateSensorById(sensor types.Sensor) error {
	args := m.Called(sensor)
	return args.Error(0)
}

func (m *MockSensorRepoForTemp) DeleteSensorByName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockSensorRepoForTemp) GetSensorByName(name string) (*types.Sensor, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Sensor), args.Error(1)
}

func (m *MockSensorRepoForTemp) SetEnabledSensorByName(name string, enabled bool) error {
	args := m.Called(name, enabled)
	return args.Error(0)
}

func (m *MockSensorRepoForTemp) GetAllSensors() ([]types.Sensor, error) {
	args := m.Called()
	return args.Get(0).([]types.Sensor), args.Error(1)
}

func (m *MockSensorRepoForTemp) GetSensorsByType(sensorType string) ([]types.Sensor, error) {
	args := m.Called(sensorType)
	return args.Get(0).([]types.Sensor), args.Error(1)
}

func (m *MockSensorRepoForTemp) GetSensorIdByName(name string) (int, error) {
	args := m.Called(name)
	return args.Int(0), args.Error(1)
}

func (m *MockSensorRepoForTemp) SensorExists(name string) (bool, error) {
	args := m.Called(name)
	return args.Bool(0), args.Error(1)
}

func (m *MockSensorRepoForTemp) UpdateSensorHealthById(sensorId int, healthStatus types.SensorHealthStatus, healthReason string) error {
	args := m.Called(sensorId, healthStatus, healthReason)
	return args.Error(0)
}

func (m *MockSensorRepoForTemp) GetSensorHealthHistoryById(sensorId int, limit int) ([]types.SensorHealthHistory, error) {
	args := m.Called(sensorId, limit)
	return args.Get(0).([]types.SensorHealthHistory), args.Error(1)
}

func (m *MockSensorRepoForTemp) DeleteHealthHistoryOlderThan(cutoffDate time.Time) error {
	args := m.Called(cutoffDate)
	return args.Error(0)
}

// ============================================================================
// Add tests
// ============================================================================

func TestTemperatureRepository_Add_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings := []types.TemperatureReading{
		{SensorName: "sensor-1", Time: "2026-01-16 12:00:00", Temperature: 22.5},
		{SensorName: "sensor-2", Time: "2026-01-16 12:00:00", Temperature: 23.0},
	}

	sensorMock.On("GetSensorIdByName", "sensor-1").Return(1, nil)
	sensorMock.On("GetSensorIdByName", "sensor-2").Return(2, nil)

	dbMock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs(1, "2026-01-16 12:00:00", "22.5").
		WillReturnResult(sqlmock.NewResult(1, 1))
	dbMock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs(2, "2026-01-16 12:00:00", "23").
		WillReturnResult(sqlmock.NewResult(2, 1))

	err := repo.Add(readings)

	assert.NoError(t, err)
	sensorMock.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_Add_EmptySlice(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings := []types.TemperatureReading{}

	err := repo.Add(readings)

	assert.NoError(t, err)
	sensorMock.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_Add_SensorNotFound(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings := []types.TemperatureReading{
		{SensorName: "nonexistent", Time: "2026-01-16 12:00:00", Temperature: 22.5},
	}

	sensorMock.On("GetSensorIdByName", "nonexistent").Return(0, errors.New("sensor not found"))

	err := repo.Add(readings)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "issue finding sensor id")
	sensorMock.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_Add_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings := []types.TemperatureReading{
		{SensorName: "sensor-1", Time: "2026-01-16 12:00:00", Temperature: 22.5},
	}

	sensorMock.On("GetSensorIdByName", "sensor-1").Return(1, nil)

	dbMock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs(1, "2026-01-16 12:00:00", "22.5").
		WillReturnError(errors.New("database error"))

	err := repo.Add(readings)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "issue persisting readings to database")
	sensorMock.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// GetBetweenDates tests
// ============================================================================

func TestTemperatureRepository_GetBetweenDates_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WithArgs("2026-01-15 00:00:00", "2026-01-16 23:59:59").
		WillReturnRows(sqlmock.NewRows(temperatureReadingColumns).
			AddRow(1, "sensor-1", "2026-01-15T10:00:00Z", 22.5).
			AddRow(2, "sensor-1", "2026-01-15T11:00:00Z", 23.0))

	readings, err := repo.GetBetweenDates(types.TableTemperatureReadings, "2026-01-15 00:00:00", "2026-01-16 23:59:59")

	assert.NoError(t, err)
	assert.Len(t, readings, 2)
	assert.Equal(t, 22.5, readings[0].Temperature)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetBetweenDates_HourlyTable(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.average_temperature FROM hourly_avg_temperature tr").
		WithArgs("2026-01-15 00:00:00", "2026-01-16 23:59:59").
		WillReturnRows(sqlmock.NewRows(temperatureReadingColumns).
			AddRow(1, "sensor-1", "2026-01-15T10:00:00Z", 22.5))

	readings, err := repo.GetBetweenDates(types.TableHourlyAverageTemperature, "2026-01-15 00:00:00", "2026-01-16 23:59:59")

	assert.NoError(t, err)
	assert.Len(t, readings, 1)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetBetweenDates_InvalidTable(t *testing.T) {
	db, _ := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings, err := repo.GetBetweenDates("invalid_table", "2026-01-15 00:00:00", "2026-01-16 23:59:59")

	assert.Error(t, err)
	assert.Nil(t, readings)
	assert.Contains(t, err.Error(), "invalid table name")
}

func TestTemperatureRepository_GetBetweenDates_EmptyRange(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WithArgs("2026-01-15 00:00:00", "2026-01-15 00:00:00").
		WillReturnRows(sqlmock.NewRows(temperatureReadingColumns))

	readings, err := repo.GetBetweenDates(types.TableTemperatureReadings, "2026-01-15 00:00:00", "2026-01-15 00:00:00")

	assert.NoError(t, err)
	assert.Empty(t, readings)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetBetweenDates_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WithArgs("2026-01-15 00:00:00", "2026-01-16 23:59:59").
		WillReturnError(errors.New("database error"))

	readings, err := repo.GetBetweenDates(types.TableTemperatureReadings, "2026-01-15 00:00:00", "2026-01-16 23:59:59")

	assert.Error(t, err)
	assert.Nil(t, readings)
	assert.Contains(t, err.Error(), "error fetching readings")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// GetTotalReadingsBySensorId tests
// ============================================================================

func TestTemperatureRepository_GetTotalReadingsBySensorId_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM temperature_readings WHERE sensor_id = \\?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(150))

	count, err := repo.GetTotalReadingsBySensorId(1)

	assert.NoError(t, err)
	assert.Equal(t, 150, count)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetTotalReadingsBySensorId_Zero(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM temperature_readings WHERE sensor_id = \\?").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	count, err := repo.GetTotalReadingsBySensorId(999)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetTotalReadingsBySensorId_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM temperature_readings WHERE sensor_id = \\?").
		WithArgs(1).
		WillReturnError(errors.New("database error"))

	count, err := repo.GetTotalReadingsBySensorId(1)

	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "error fetching total readings")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// GetLatest tests
// ============================================================================

func TestTemperatureRepository_GetLatest_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WillReturnRows(sqlmock.NewRows(temperatureReadingColumns).
			AddRow(1, "sensor-1", "2026-01-16T12:00:00Z", 22.5).
			AddRow(2, "sensor-2", "2026-01-16T12:00:00Z", 23.0).
			AddRow(3, "sensor-1", "2026-01-16T11:00:00Z", 21.5)) // older reading for sensor-1, should be deduplicated

	readings, err := repo.GetLatest()

	assert.NoError(t, err)
	assert.Len(t, readings, 2) // Should deduplicate to latest per sensor
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetLatest_SingleSensor(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WillReturnRows(sqlmock.NewRows(temperatureReadingColumns).
			AddRow(1, "sensor-1", "2026-01-16T12:00:00Z", 22.5))

	readings, err := repo.GetLatest()

	assert.NoError(t, err)
	assert.Len(t, readings, 1)
	assert.Equal(t, "sensor-1", readings[0].SensorName)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetLatest_EmptyTable(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WillReturnRows(sqlmock.NewRows(temperatureReadingColumns))

	readings, err := repo.GetLatest()

	assert.NoError(t, err)
	assert.Empty(t, readings)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetLatest_DBError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT tr.id, s.name AS sensor_name, tr.time, tr.temperature FROM temperature_readings tr").
		WillReturnError(errors.New("database error"))

	readings, err := repo.GetLatest()

	assert.Error(t, err)
	assert.Nil(t, readings)
	assert.Contains(t, err.Error(), "error fetching latest readings")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// DeleteReadingsOlderThan tests
// ============================================================================

func TestTemperatureRepository_DeleteReadingsOlderThan_Success(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	cutoff := time.Now().Add(-24 * time.Hour)

	dbMock.ExpectBegin()
	dbMock.ExpectExec("DELETE FROM temperature_readings WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnResult(sqlmock.NewResult(0, 100))
	dbMock.ExpectExec("DELETE FROM hourly_avg_temperature WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnResult(sqlmock.NewResult(0, 50))
	dbMock.ExpectCommit()

	err := repo.DeleteReadingsOlderThan(cutoff)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_DeleteReadingsOlderThan_NothingToDelete(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	cutoff := time.Now().Add(-24 * time.Hour)

	dbMock.ExpectBegin()
	dbMock.ExpectExec("DELETE FROM temperature_readings WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectExec("DELETE FROM hourly_avg_temperature WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnResult(sqlmock.NewResult(0, 0))
	dbMock.ExpectCommit()

	err := repo.DeleteReadingsOlderThan(cutoff)

	assert.NoError(t, err)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_DeleteReadingsOlderThan_RollbackOnFirstDeleteError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	cutoff := time.Now().Add(-24 * time.Hour)

	dbMock.ExpectBegin()
	dbMock.ExpectExec("DELETE FROM temperature_readings WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnError(errors.New("database error"))
	dbMock.ExpectRollback()

	err := repo.DeleteReadingsOlderThan(cutoff)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting old temperature readings")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_DeleteReadingsOlderThan_RollbackOnSecondDeleteError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	cutoff := time.Now().Add(-24 * time.Hour)

	dbMock.ExpectBegin()
	dbMock.ExpectExec("DELETE FROM temperature_readings WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnResult(sqlmock.NewResult(0, 100))
	dbMock.ExpectExec("DELETE FROM hourly_avg_temperature WHERE time < \\?").
		WithArgs(cutoff).
		WillReturnError(errors.New("database error"))
	dbMock.ExpectRollback()

	err := repo.DeleteReadingsOlderThan(cutoff)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error deleting old temperature readings")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_DeleteReadingsOlderThan_BeginTxError(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	cutoff := time.Now().Add(-24 * time.Hour)

	dbMock.ExpectBegin().WillReturnError(errors.New("connection error"))

	err := repo.DeleteReadingsOlderThan(cutoff)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

// ============================================================================
// Edge case tests
// ============================================================================

func TestTemperatureRepository_Add_NegativeTemperature(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings := []types.TemperatureReading{
		{SensorName: "sensor-1", Time: "2026-01-16 12:00:00", Temperature: -15.5},
	}

	sensorMock.On("GetSensorIdByName", "sensor-1").Return(1, nil)

	dbMock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs(1, "2026-01-16 12:00:00", "-15.5").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Add(readings)

	assert.NoError(t, err)
	sensorMock.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_Add_ZeroTemperature(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	readings := []types.TemperatureReading{
		{SensorName: "sensor-1", Time: "2026-01-16 12:00:00", Temperature: 0.0},
	}

	sensorMock.On("GetSensorIdByName", "sensor-1").Return(1, nil)

	dbMock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs(1, "2026-01-16 12:00:00", "0").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Add(readings)

	assert.NoError(t, err)
	sensorMock.AssertExpectations(t)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetTotalReadingsBySensorId_NegativeId(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	// Negative ID should still work at the SQL level, just return 0
	dbMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM temperature_readings WHERE sensor_id = \\?").
		WithArgs(-1).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	count, err := repo.GetTotalReadingsBySensorId(-1)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}

func TestTemperatureRepository_GetTotalReadingsBySensorId_ZeroId(t *testing.T) {
	db, dbMock := newMockDB(t)
	sensorMock := new(MockSensorRepoForTemp)
	repo := NewTemperatureRepository(db, sensorMock)

	dbMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM temperature_readings WHERE sensor_id = \\?").
		WithArgs(0).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	count, err := repo.GetTotalReadingsBySensorId(0)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
