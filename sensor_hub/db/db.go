package database

import (
	"database/sql"
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/types"
	"example/sensorHub/utils"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
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
			return fmt.Errorf("issue persisting readings to database: %w", err)
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
			log.Printf("Error scanning row, skipping: %v", err)
			continue
		}
		readings = append(readings, reading)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
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
			log.Printf("Error scanning row, skipping: %v", err)
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
		return nil, fmt.Errorf("error iterating over rows: %w", err)
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

// This function initialises the database connection and creates the sensor_database if it does not exist.
// It expects the database properties to be provided through a map with keys:
// "database.username", "database.password", "database.hostname", and "database.port".
func InitialiseDatabase() error {
	db_username := appProps.DATABASE_PROPERTIES["database.username"]
	db_password := appProps.DATABASE_PROPERTIES["database.password"]
	db_hostname := appProps.DATABASE_PROPERTIES["database.hostname"]
	db_port := appProps.DATABASE_PROPERTIES["database.port"]

	jdbc_url := db_username + ":" + db_password + "@(" + db_hostname + ":" + db_port + ")/sensor_database?parseTime=true"

	db, err := sql.Open("mysql", jdbc_url)
	if err != nil {
		return fmt.Errorf("could not initialise connection to database: %w", err)
	}
	DB = db
	log.Println("Connected to database")

	return nil
}
