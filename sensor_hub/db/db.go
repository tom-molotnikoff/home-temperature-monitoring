package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"

	appProps "example/sensorHub/application_properties"
	"example/sensorHub/types"
	"example/sensorHub/utils"
)

var DB *sql.DB

const (
	TableTemperatureReadings      = "temperature_readings"
	TableHourlyAverageTemperature = "hourly_avg_temperature"
)

// This function adds a list of sensor readings to the temperature_readings table in the sensor_database.
// It expects the readings to be in the form of a slice of SensorReading pointers.
// Each SensorReading should have a Name and a Reading field, where Reading is a struct containing
// Temperature (float64) and Time (string).
// It will log an error if there is an issue persisting the readings to the database.
func AddListOfRawReadings(readings []types.APIReading) error {
	convertedDbReadings := utils.ConvertAPIReadingsToDbReadings(readings)
	query := fmt.Sprintf("INSERT INTO %s (sensor_name, time, temperature) VALUES (?, ?, ?)", TableTemperatureReadings)
	for _, reading := range convertedDbReadings {
		_, err := DB.Exec(query, reading.SensorName, reading.Time, strconv.FormatFloat(reading.Temperature, 'f', -1, 64))
		if err != nil {
			return fmt.Errorf("issue persisting readings to database: %s", err)
		}
		log.Printf("Saved a reading from Sensor(%s) into the database", reading.SensorName)
	}
	return nil
}

// This function will fetch readings from the database between the specified start and end dates.
// It will log the readings or any errors encountered during the process.
var GetReadingsBetweenDates = func(tableName string, startDate string, endDate string) ([]types.APIReading, error) {
	if tableName != TableTemperatureReadings && tableName != TableHourlyAverageTemperature {
		return nil, fmt.Errorf("invalid table name: %s", tableName)
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE time BETWEEN ? AND ? ORDER BY time ASC", tableName)

	rows, err := DB.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error fetching readings between %s and %s: %w", startDate, endDate, err)
	}
	defer rows.Close()
	var readings []types.DbReading
	for rows.Next() {
		var reading types.DbReading

		err := rows.Scan(&reading.Id, &reading.SensorName, &reading.Time, &reading.Temperature)
		if err != nil {
			log.Printf("Error scanning row: %s", err)
			continue
		}
		readings = append(readings, reading)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %s", err)
		return nil, fmt.Errorf("error iterating over rows: %s", err)
	}
	apiReadings := utils.ConvertDbReadingsToApiReadings(readings)
	return apiReadings, nil
}

// This function retrieves the latest readings from the temperature_readings table in the sensor_database.
// It will only return the first occurrence of each sensor name, ensuring that only the latest reading
// for each sensor is included in the result. If a sensor hasn't been read in the last 30 readings, it won't be included.
var GetLatestReadings = func() ([]types.APIReading, error) {
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY time DESC LIMIT 30", TableTemperatureReadings)
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest readings: %w", err)
	}
	defer rows.Close()
	var readings []types.DbReading

	for rows.Next() {
		var reading types.DbReading
		err := rows.Scan(&reading.Id, &reading.SensorName, &reading.Time, &reading.Temperature)
		if err != nil {
			log.Printf("Error scanning row: %s", err)
			continue
		}
		readings = append(readings, reading)
	}

	// Use a map to track the latest reading per sensor
	latestReadingsPerSensor := make(map[string]types.DbReading)
	for _, r := range readings {
		sensorName := r.SensorName
		// Only add if not already present (first occurrence is the latest due to DESC order)
		if _, exists := latestReadingsPerSensor[sensorName]; !exists {
			latestReadingsPerSensor[sensorName] = types.DbReading{
				Id:          r.Id,
				SensorName:  r.SensorName,
				Time:        r.Time,
				Temperature: r.Temperature,
			}
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %s", err)
		return nil, fmt.Errorf("error iterating over rows: %s", err)
	}
	// Copy map values to slice
	finalReadings := make([]types.DbReading, 0, len(latestReadingsPerSensor))
	for _, r := range latestReadingsPerSensor {
		finalReadings = append(finalReadings, r)
	}

	// Convert to APIReading format for consistency
	apiReadings := utils.ConvertDbReadingsToApiReadings(finalReadings)
	return apiReadings, nil
}

// This function creates the temperature_readings table in the sensor_database if it does not exist.
func createTemperatureReadingsTable() error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT,
			sensor_name TEXT NOT NULL,
			time DATETIME NOT NULL,
			temperature FLOAT(4) NOT NULL,
			PRIMARY KEY (id)
		);`, TableTemperatureReadings)

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("issue creating temperature readings table: %s", err)
	}
	// Create indexes separately, without IF NOT EXISTS
	_, err = DB.Exec(`CREATE INDEX hourly_idx_time ON temperature_readings (time DESC);`)
	if err != nil {
		log.Printf("Could not create hourly_idx_time index (may already exist): %s", err)
	}
	_, err = DB.Exec(`CREATE INDEX hourly_idx_sensor_name ON temperature_readings (sensor_name(16));`)
	if err != nil {
		log.Printf("Could not create hourly_idx_sensor_name index (may already exist): %s", err)
	}

	return nil
}

// This function creates the hourly_average_temperature table and the associated event
// in the sensor_database if they do not exist. The event calculates the hourly average temperature
// for each sensor and inserts it into the hourly_average_temperature table every hour.
// It ensures that duplicate entries for the same sensor and hour are not created.
func createEventForHourlyAverageTemperature() error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
			id INT AUTO_INCREMENT,
			sensor_name VARCHAR(16) NOT NULL,
			time DATETIME NOT NULL,
			average_temperature FLOAT(4) NOT NULL,
			PRIMARY KEY (id),
			UNIQUE KEY unique_sensor_hour (sensor_name, time)
		);`, TableHourlyAverageTemperature)

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("issue creating %s table: %s", TableHourlyAverageTemperature, err)
	}

	query = fmt.Sprintf(`CREATE INDEX idx_time ON %s (time DESC);`, TableHourlyAverageTemperature)
	_, err = DB.Exec(query)
	if err != nil {
		log.Printf("%s: Could not create idx_time index (may already exist): %s", TableHourlyAverageTemperature, err)
	}
	query = fmt.Sprintf(`CREATE INDEX idx_sensor_name ON %s (sensor_name(16));`, TableHourlyAverageTemperature)
	_, err = DB.Exec(query)
	if err != nil {
		log.Printf("%s: Could not create idx_sensor_name index (may already exist): %s", TableHourlyAverageTemperature, err)
	}

	_, err = DB.Exec("DROP EVENT IF EXISTS hourly_average_temperature_event;")
	if err != nil {
		log.Printf("Could not drop existing event (most likely it does not exist): %s", err)
	}

	query = `
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
	`

	_, err = DB.Exec(query)
	if err != nil {
		return fmt.Errorf("issue creating hourly average temperature event: %s", err)
	}
	return nil
}

// This function initialises the database connection and creates the sensor_database if it does not exist.
// It expects the database properties to be provided through a map with keys:
// "database.username", "database.password", "database.hostname", and "database.port".
func InitialiseDatabase() error {
	db_username := appProps.DATABASE_PROPERTIES["database.username"]
	db_password := appProps.DATABASE_PROPERTIES["database.password"]
	db_hostname := appProps.DATABASE_PROPERTIES["database.hostname"]
	db_port := appProps.DATABASE_PROPERTIES["database.port"]

	jdbc_url := db_username + ":" + db_password + "@(" + db_hostname + ":" + db_port + ")/?parseTime=true"

	db, err := sql.Open("mysql", jdbc_url)

	if err != nil {
		return fmt.Errorf("could not initialise connection to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("could not ping database: %w", err)
	}

	create_database_query := "CREATE DATABASE IF NOT EXISTS sensor_database"

	_, err = db.Exec(create_database_query)

	if err != nil {
		return fmt.Errorf("could not create database: %s", err)
	}

	db.Close()

	jdbc_url = db_username + ":" + db_password + "@(" + db_hostname + ":" + db_port + ")/sensor_database?parseTime=true"

	db, err = sql.Open("mysql", jdbc_url)
	if err != nil {
		return fmt.Errorf("could not initialise connection to database: %s", err)
	}
	DB = db
	log.Println("Connected to database")

	err = createTemperatureReadingsTable()
	if err != nil {
		return fmt.Errorf("failed to create temperature readings table: %w", err)
	}
	err = createEventForHourlyAverageTemperature()
	if err != nil {
		return fmt.Errorf("failed to create hourly average temperature table and event: %w", err)
	}

	return nil
}
