package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type APIReading struct {
	SensorName string `json:"sensor_name"`
	Reading    struct {
		Temperature float64 `json:"temperature"`
		Time        string  `json:"time"`
	} `json:"reading"`
}

type Reading struct {
	Id          int     `json:"id"`
	SensorName  string  `json:"sensor_name"`
	Time        string  `json:"time"`
	Temperature float64 `json:"temperature"`
}

// This function adds a list of sensor readings to the temperature_readings table in the sensor_database.
// It expects the readings to be in the form of a slice of SensorReading pointers.
// Each SensorReading should have a Name and a Reading field, where Reading is a struct containing
// Temperature (float64) and Time (string).
// It will log an error if there is an issue persisting the readings to the database.
func add_list_of_readings(readings []*SensorReading) error {
	for _, reading := range readings {
		_, err := DB.Exec(`INSERT INTO temperature_readings (sensor_name, time, temperature) VALUES (?, ?, ?)`, reading.SensorName, reading.Reading.Time, strconv.FormatFloat(reading.Reading.Temperature, 'f', -1, 64))
		if err != nil {
			return fmt.Errorf("issue persisting readings to database: %s", err)
		}
		log.Printf("Saved a reading from Sensor(%s) into the database", reading.SensorName)
	}
	return nil
}

// This function will fetch readings from the database between the specified start and end dates.
// It will log the readings or any errors encountered during the process.
func getReadingsBetweenDates(startDate string, endDate string) (*[]APIReading, error) {

	query := "SELECT * FROM temperature_readings WHERE time BETWEEN ? AND ?"
	rows, err := DB.Query(query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("error fetching readings between %s and %s: %w", startDate, endDate, err)
	}
	defer rows.Close()
	var readings []Reading
	for rows.Next() {
		var reading Reading

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
	var apiReadings []APIReading
	for _, r := range readings {
		apiReadings = append(apiReadings, APIReading{
			SensorName: r.SensorName,
			Reading: struct {
				Temperature float64 `json:"temperature"`
				Time        string  `json:"time"`
			}{
				Temperature: math.Round(r.Temperature*10) / 10,
				Time:        r.Time,
			},
		})
	}
	return &apiReadings, nil
}

// This function validates the database properties by checking if the required fields are set.
func validateDatabaseProperties() error {
	if DATABASE_PROPERTIES["database.username"] == "" || DATABASE_PROPERTIES["database.password"] == "" ||
		DATABASE_PROPERTIES["database.hostname"] == "" || DATABASE_PROPERTIES["database.port"] == "" {
		return fmt.Errorf("database properties are not set correctly. please check your database.properties file")
	}
	return nil
}

// This function retrieves the latest readings from the temperature_readings table in the sensor_database.
// It will only return the first occurrence of each sensor name, ensuring that only the latest reading
// for each sensor is included in the result. If a sensor hasn't been read in the last 30 readings, it won't be included.
func getLatestReadings() ([]APIReading, error) {
	query := "SELECT sensor_name, time, temperature FROM temperature_readings ORDER BY time DESC LIMIT 30"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest readings: %w", err)
	}
	defer rows.Close()
	latest := make(map[string]APIReading)
	for rows.Next() {
		var sensorName, timeStr string
		var temperature float64
		err := rows.Scan(&sensorName, &timeStr, &temperature)
		if err != nil {
			log.Printf("Error scanning row: %s", err)
			continue
		}
		// Only add if not already present (first occurrence is the latest due to DESC order)
		if _, exists := latest[sensorName]; !exists {
			latest[sensorName] = APIReading{
				SensorName: sensorName,
				Reading: struct {
					Temperature float64 `json:"temperature"`
					Time        string  `json:"time"`
				}{
					Temperature: math.Round(temperature*10) / 10,
					Time:        timeStr,
				},
			}
		}
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %s", err)
		return nil, fmt.Errorf("error iterating over rows: %s", err)
	}
	// Copy map values to slice
	readings := make([]APIReading, 0, len(latest))
	for _, r := range latest {
		readings = append(readings, r)
	}
	return readings, nil
}

// This function creates the temperature_readings table in the sensor_database if it does not exist.
func create_temperature_readings_table() error {
	query := `
		CREATE TABLE IF NOT EXISTS temperature_readings (
			id INT AUTO_INCREMENT,
			sensor_name TEXT NOT NULL,
			time DATETIME NOT NULL,
			temperature FLOAT(4) NOT NULL,
			PRIMARY KEY (id)
		);
	`
	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("issue creating temperature readings table: %s", err)
	}
	// Create indexes separately, without IF NOT EXISTS
	_, err = DB.Exec(`CREATE INDEX idx_time ON temperature_readings (time DESC);`)
	if err != nil {
		log.Printf("Could not create idx_time index (may already exist): %s", err)
	}
	_, err = DB.Exec(`CREATE INDEX idx_sensor_name ON temperature_readings (sensor_name(16));`)
	if err != nil {
		log.Printf("Could not create idx_sensor_name index (may already exist): %s", err)
	}

	return nil
}

// This function initialises the database connection and creates the sensor_database if it does not exist.
// It expects the database properties to be provided through a map with keys:
// "database.username", "database.password", "database.hostname", and "database.port".
func initialise_database(db_properties map[string]string) (*sql.DB, error) {
	db_username := db_properties["database.username"]
	db_password := db_properties["database.password"]
	db_hostname := db_properties["database.hostname"]
	db_port := db_properties["database.port"]

	jdbc_url := db_username + ":" + db_password + "@(" + db_hostname + ":" + db_port + ")/?parseTime=true"

	db, err := sql.Open("mysql", jdbc_url)

	if err != nil {
		return nil, fmt.Errorf("could not initialise connection to database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("could not ping database: %w", err)
	}

	create_database_query := "CREATE DATABASE IF NOT EXISTS sensor_database"

	_, err = db.Exec(create_database_query)

	if err != nil {
		return nil, fmt.Errorf("could not create database: %s", err)
	}

	db.Close()

	jdbc_url = db_username + ":" + db_password + "@(" + db_hostname + ":" + db_port + ")/sensor_database?parseTime=true"

	db, err = sql.Open("mysql", jdbc_url)
	if err != nil {
		return nil, fmt.Errorf("could not initialise connection to database: %s", err)
	}

	log.Println("Connected to database")
	return db, nil
}
