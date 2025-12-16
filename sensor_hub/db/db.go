package database

import (
	"database/sql"
	appProps "example/sensorHub/application_properties"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func InitialiseDatabase() (*sql.DB, error) {
	dbUsername := appProps.DatabaseProperties["database.username"]
	dbPassword := appProps.DatabaseProperties["database.password"]
	dbHostname := appProps.DatabaseProperties["database.hostname"]
	dbPort := appProps.DatabaseProperties["database.port"]

	jdbcUrl := dbUsername + ":" + dbPassword + "@(" + dbHostname + ":" + dbPort + ")/sensor_database?parseTime=true"

	db, err := sql.Open("mysql", jdbcUrl)
	if err != nil {
		return nil, fmt.Errorf("could not initialise connection to database: %w", err)
	}
	log.Println("Connected to database")

	return db, nil
}
