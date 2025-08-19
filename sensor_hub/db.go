package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// This function adds a list of sensor readings to the temperature_readings table in the sensor_database.
// It expects the readings to be in the form of a slice of SensorReading pointers.
// Each SensorReading should have a Name and a Reading field, where Reading is a struct containing
// Temperature (float64) and Time (string).
// It will log an error if there is an issue persisting the readings to the database.
func add_list_of_readings(readings []*SensorReading) error {
	for _, reading := range readings {
		_, err := DB.Exec(`INSERT INTO temperature_readings (sensor_name, time, temperature) VALUES (?, ?, ?)`, reading.Name, reading.Reading.Time, strconv.FormatFloat(reading.Reading.Temperature, 'f', -1, 64))
		if err != nil {
			return fmt.Errorf("issue persisting readings to database: %s", err)
		}
		log.Printf("Saved a reading from Sensor(%s) into the database", reading.Name)
	}
	return nil
}

func validateDatabaseProperties() error {
	if DATABASE_PROPERTIES["database.username"] == "" || DATABASE_PROPERTIES["database.password"] == "" ||
		DATABASE_PROPERTIES["database.hostname"] == "" || DATABASE_PROPERTIES["database.port"] == "" {
		return fmt.Errorf("database properties are not set correctly. please check your database.properties file")
	}
	return nil
}

// This function selects the last two readings from the temperature_readings table in the sensor_database
// and logs them. It assumes that the table has been created and populated with readings.
// It will log an error if there is an issue fetching the data or scanning the rows.
func logLast2Readings() error {
	query := "SELECT * FROM temperature_readings ORDER BY time DESC LIMIT 2;"

	type Reading struct {
		id          int
		sensor_name string
		time        string
		temperature float64
	}

	rows, err := DB.Query(query)

	if err != nil {
		return fmt.Errorf("there was an error fetching the readings from the database: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reading Reading
		err := rows.Scan(&reading.id, &reading.sensor_name, &reading.time, &reading.temperature)

		if err != nil {
			return fmt.Errorf("there was an error scanning the rows from the results of the query: %s", err)
		}
		log.Printf("FROM DATABASE: Sensor: %s, Time: %s, Temperature: %s", reading.sensor_name, reading.time, strconv.FormatFloat(reading.temperature, 'f', -1, 64))
	}
	return nil
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
