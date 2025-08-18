package main

import (
	"bufio"
	"database/sql"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func add_list_of_readings(readings []*SensorReading) {
	for _, reading := range readings {
		_, err := DB.Exec(`INSERT INTO temperature_readings (sensor_name, time, temperature) VALUES (?, ?, ?)`, reading.Name, reading.Reading.Time, strconv.FormatFloat(reading.Reading.Temperature, 'f', -1, 64))
		if err != nil {
			log.Fatalf("Issue persisting readings to database: %s", err)
		}
		log.Printf("Saved a reading from Sensor(%s) into the database", reading.Name)
	}

}

func read_db_properties_file(path string) map[string]string {
	file, err := os.Open("database.properties")
	if err != nil {
		log.Fatalf("Failed to read database.properties: %s", err)
	}
	defer file.Close()

	db_properties := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			db_properties[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read database.properties: %s", err)
	}
	return db_properties
}

func selectLast2Readings() {
	query := "SELECT * FROM temperature_readings ORDER BY time DESC LIMIT 2;"

	type Reading struct {
		id          int
		sensor_name string
		time        string
		temperature float64
	}

	rows, err := DB.Query(query)

	if err != nil {
		log.Printf("There was an error fetching from the database: %s", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var reading Reading
		err := rows.Scan(&reading.id, &reading.sensor_name, &reading.time, &reading.temperature)

		if err != nil {
			log.Printf("Something went wrong reading the data from the pets table: %s", err)
			return
		}
		log.Printf("FROM DATABASE: Sensor: %s, Time: %s, Temperature: %s", reading.sensor_name, reading.time, strconv.FormatFloat(reading.temperature, 'f', -1, 64))
	}
}

func create_temperature_readings_table() {
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
		log.Fatalf("Issue creating temperature readings table: %s", err)
	}
}

// FIXME: need to consume the SQL database properties from a git ignored config file !
func initialise_database(db_properties map[string]string) *sql.DB {
	db_username := db_properties["database.username"]
	db_password := db_properties["database.password"]
	db_hostname := db_properties["database.hostname"]
	db_port := db_properties["database.port"]

	jdbc_url := db_username + ":" + db_password + "@(" + db_hostname + ":" + db_port + ")/?parseTime=true"

	db, err := sql.Open("mysql", jdbc_url)

	if err != nil {
		log.Fatalf("Could not initialise connection to database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not initialise connection to database: %s", err)
	}

	create_database_query := "CREATE DATABASE IF NOT EXISTS sensor_database"

	_, err = db.Exec(create_database_query)

	if err != nil {
		log.Fatalf("Could not create go_db: %s", err)
	}

	db.Close()

	db, err = sql.Open("mysql", "root:password@(localhost:3306)/go_db?parseTime=true")
	if err != nil {
		log.Fatalf("Could not initialise connection to database: %s", err)
	}

	log.Println("Connected to database")
	return db
}
