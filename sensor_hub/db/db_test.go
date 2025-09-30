package database

import (
	"example/sensorHub/types"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestAddListOfRawReadings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	
	mock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs("sensor1","2025-10-01T10:00:00Z", "23.5").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs("sensor2","2025-10-01T11:00:00Z", "25").
		WillReturnResult(sqlmock.NewResult(2, 1))
	readings := []types.APIReading{
		{SensorName: "sensor1", Reading: types.RawTemperatureReading{Temperature: 23.5, Time: "2025-10-01T10:00:00Z"}},
		{SensorName: "sensor2", Reading: types.RawTemperatureReading{Temperature: 25.0, Time: "2025-10-01T11:00:00Z"}},
	}
	err = AddListOfRawReadings(readings)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestAddListOfRawReadings_EmptyList(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db

	readings := []types.APIReading{}
	err = AddListOfRawReadings(readings)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAddListOfRawReadings_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db

	mock.ExpectExec("INSERT INTO temperature_readings").
		WithArgs("sensor1","2025-10-01T10:00:00Z", "23.5").
		WillReturnError(sqlmock.ErrCancelled)

	readings := []types.APIReading{
		{SensorName: "sensor1", Reading: types.RawTemperatureReading{Temperature: 23.5, Time: "2025-10-01T10:00:00Z"}},
	}
	err = AddListOfRawReadings(readings)

	if err == nil {
		t.Error("expected an error but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetReadingsBetweenDates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM temperature_readings WHERE time BETWEEN ? AND ? ORDER BY time ASC")).
		WithArgs("2025-10-01T09:00:00Z", "2025-10-01T12:00:00Z").
		WillReturnRows(sqlmock.NewRows([]string{"id", "sensor_name", "time", "temperature"}).
			AddRow(1, "sensor1", "2025-10-01T10:00:00Z", 23.5).
			AddRow(2, "sensor2", "2025-10-01T11:00:00Z", 25.0))

	readings, err := GetReadingsBetweenDates("temperature_readings","2025-10-01T09:00:00Z", "2025-10-01T12:00:00Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(readings) != 2 {
		t.Errorf("expected 2 readings but got %d", len(readings))
	}
	expectedFirst := types.APIReading{SensorName: "sensor1", Reading: types.RawTemperatureReading{Temperature: 23.5, Time: "2025-10-01T10:00:00Z"}}
	if readings[0] != expectedFirst {
		t.Errorf("expected first reading to be %v but got %v", expectedFirst, readings[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetReadingsBetweenDates_InvalidStartDate(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	
	_, err = GetReadingsBetweenDates("temperature_readings","invalid-date", "2025-10-01T12:00:00Z")
	if err == nil {
		t.Error("expected an error but got none")
	}
}

func TestGetReadingsBetweenDates_InvalidEndDate(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	
	_, err = GetReadingsBetweenDates("temperature_readings","2025-10-01T09:00:00Z", "invalid-date")
	if err == nil {
		t.Error("expected an error but got none")
	}
}

func TestGetReadingsBetweenDates_InvalidTableName(t *testing.T){
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	
	_, err = GetReadingsBetweenDates("invalid_table","2025-10-01T09:00:00Z", "2025-10-01T12:00:00Z")
	if err == nil {
		t.Error("expected an error but got none")
	}
}

// Essentially an SQL contract test to ensure the SQL statements are correct
func TestCreateTemperatureReadingsTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	mock.ExpectExec(regexp.QuoteMeta(`CREATE TABLE IF NOT EXISTS temperature_readings (
			id INT AUTO_INCREMENT,
			sensor_name TEXT NOT NULL,
			time DATETIME NOT NULL,
			temperature FLOAT(4) NOT NULL,
			PRIMARY KEY (id)
		);`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`CREATE INDEX hourly_idx_time ON temperature_readings (time DESC);`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta(`CREATE INDEX hourly_idx_sensor_name ON temperature_readings (sensor_name(16));`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
		
	err = createTemperatureReadingsTable()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

// Essentially an SQL contract test to ensure the SQL statements are correct
func TestCreateEventForHourlyAverageTemperature(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db
	mock.ExpectExec(regexp.QuoteMeta("CREATE TABLE IF NOT EXISTS hourly_avg_temperature ( id INT AUTO_INCREMENT, sensor_name VARCHAR(16) NOT NULL, time DATETIME NOT NULL, average_temperature FLOAT(4) NOT NULL, PRIMARY KEY (id), UNIQUE KEY unique_sensor_hour (sensor_name, time) );")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX idx_time ON hourly_avg_temperature (time DESC);")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("CREATE INDEX idx_sensor_name ON hourly_avg_temperature (sensor_name(16));")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta("DROP EVENT IF EXISTS hourly_average_temperature_event;")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(regexp.QuoteMeta(`
			CREATE EVENT IF NOT EXISTS hourly_average_temperature_event
			ON SCHEDULE EVERY 1 HOUR
			STARTS TIMESTAMP(CURRENT_DATE, SEC_TO_TIME((HOUR(NOW())+1)*3600 + 60))
			DO
				INSERT INTO hourly_avg_temperature (sensor_name, time, average_temperature)
				SELECT
						tr.sensor_name,
						DATE_FORMAT(tr.time, '%Y-%m-%d %H:00:00') AS hour,
						ROUND(AVG(tr.temperature), 2) AS avg_temp
				FROM temperature_readings tr
				WHERE tr.time >= DATE_FORMAT(DATE_SUB(NOW(), INTERVAL 1 HOUR), '%Y-%m-%d %H:00:00')
				AND tr.time < DATE_FORMAT(NOW(), '%Y-%m-%d %H:00:00')
				GROUP BY tr.sensor_name, hour
				HAVING NOT EXISTS (
						SELECT 1
						FROM hourly_avg_temperature hat
						WHERE hat.sensor_name = tr.sensor_name
							AND hat.time = hour
				);
		`)).WillReturnResult(sqlmock.NewResult(1, 1))
	err = createEventForHourlyAverageTemperature()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetLatestReadings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM temperature_readings ORDER BY time DESC LIMIT 30")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sensor_name", "time", "temperature"}).
			AddRow(1, "sensor1", "2025-10-01T10:00:00Z", 23.5).
			AddRow(2, "sensor2", "2025-10-01T11:00:00Z", 25.0).
			AddRow(3, "sensor1", "2025-10-01T09:00:00Z", 22.0)) // Older reading for sensor1

	readings, err := GetLatestReadings()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(readings) != 2 {
		t.Errorf("expected 2 unique sensor readings but got %d", len(readings))
	}
	expectedFirst := types.APIReading{SensorName: "sensor1", Reading: types.RawTemperatureReading{Temperature: 23.5, Time: "2025-10-01T10:00:00Z"}}
	if readings[0] != expectedFirst {
		t.Errorf("expected first reading to be %v but got %v", expectedFirst, readings[0])
	}
	expectedSecond := types.APIReading{SensorName: "sensor2", Reading: types.RawTemperatureReading{Temperature: 25.0, Time: "2025-10-01T11:00:00Z"}}
	if readings[1] != expectedSecond {
		t.Errorf("expected second reading to be %v but got %v", expectedSecond, readings[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetLatestReadings_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM temperature_readings ORDER BY time DESC LIMIT 30")).
		WillReturnError(sqlmock.ErrCancelled)

	_, err = GetLatestReadings()
	if err == nil {
		t.Error("expected an error but got none")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetLatestReadings_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%v' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	DB = db

	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM temperature_readings ORDER BY time DESC LIMIT 30")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "sensor_name", "time", "temperature"})) // No rows

	readings, err := GetLatestReadings()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(readings) != 0 {
		t.Errorf("expected 0 readings but got %d", len(readings))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}