package database

import (
	"database/sql"
	appProps "example/sensorHub/application_properties"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func InitialiseDatabase() (*sql.DB, error) {
	dbUsername := appProps.AppConfig.DatabaseUsername
	dbPassword := appProps.AppConfig.DatabasePassword
	dbHostname := appProps.AppConfig.DatabaseHostname
	dbPort := appProps.AppConfig.DatabasePort

	jdbcUrl := dbUsername + ":" + dbPassword + "@(" + dbHostname + ":" + dbPort + ")/sensor_database?parseTime=true"

	db, err := sql.Open("mysql", jdbcUrl)
	if err != nil {
		return nil, fmt.Errorf("could not initialise connection to database: %w", err)
	}
	log.Println("Connected to database")

	return db, nil
}
