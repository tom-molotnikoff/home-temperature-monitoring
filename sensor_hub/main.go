package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var APPLICATION_PROPERTIES map[string]string
var DATABASE_PROPERTIES map[string]string
var SMTP_PROPERTIES map[string]string

type Servers struct {
	Servers []ServerItem `yaml:"servers"`
}

type ServerItem struct {
	Url string `yaml:"url"`
}

type SensorReading struct {
	Name    string
	Reading TemperatureReading
}

type TemperatureReading struct {
	Temperature float64
	Time        string
}

// This function reads the OpenAPI specification file (openapi.yaml) to discover the URLs of temperature sensors.
// It expects the file to be in the same directory as the executable or at the specified relative path.
// It will log a fatal error if it cannot read the file or parse it correctly.
func discover_sensor_urls() ([]string, error) {
	fileData, err := os.ReadFile(APPLICATION_PROPERTIES["openapi.yaml.location"])
	if err != nil {
		log.Printf("Cannot find the openapi.yaml file for the temperature sensors: %s\n", err)
		return nil, err
	}
	var servers Servers

	err = yaml.Unmarshal(fileData, &servers)
	if err != nil {
		log.Printf("Cannot unmarshal the yaml into a map: %s\n", err)
		return nil, err
	}
	urls := make([]string, 0)

	for _, value := range servers.Servers {
		urls = append(urls, value.Url)
	}
	return urls, nil
}

// This function takes a list of sensor URLs, fetches the temperature readings from each sensor,
// and returns a slice of SensorReading objects containing the sensor name and its reading.
// It logs any errors encountered during the fetching or decoding process.
// It assumes that the sensor API returns a JSON object with the structure:
//
//	{
//	  "temperature": <float>,
//	  "time": <string>
//	}
func take_readings(servers []string) []*SensorReading {
	responses := make([]*SensorReading, 0)

	for _, server := range servers {
		readingUrl := server + "temperature"
		resp, err := http.Get(readingUrl)
		if err != nil {
			log.Printf("Issue fetching temperature from a sensor: %s\n", err)
			continue
		}
		defer resp.Body.Close()
		response := new(SensorReading)
		err = json.NewDecoder(resp.Body).Decode(response)

		if err != nil {
			log.Printf("Issue reading request body: %s\n", err)
			continue
		}

		responses = append(responses, response)
	}
	return responses
}

// This function reads a properties file and returns a map of key-value pairs
// It expects the properties file to be in the format: key=value
// It will log a fatal error if it cannot read the file or parse it correctly.
func read_properties_file(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Failed to read %s: %s", path, err)
		return nil, err
	}
	defer file.Close()

	props := make(map[string]string)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			props[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Failed to read %s: %s", path, err)
		return nil, err
	}
	return props, nil
}

// This application will read the temperature from sensors through their APIs
// persist the readings to a database, and send an email alert if the
// temperature exceeds a threshold defined in the application properties.
func main() {

	log.SetPrefix("sensor-hub: ")

	db_props, err := read_properties_file("database.properties")
	if err != nil {
		log.Fatalf("Failed to read database properties file: %s", err)
	}
	DATABASE_PROPERTIES = db_props
	err = validateDatabaseProperties()
	if err != nil {
		log.Fatalf("Database properties are not set correctly: %s", err)
	}

	SMTP_PROPERTIES, err = read_properties_file("smtp.properties")
	if err != nil {
		log.Printf("SMTP properties file missing, email alerts will be disabled: %s", err)
	} else {
		validationErr := validateSMTPProperties()
		if validationErr != nil {
			log.Printf("SMTP properties are present but not set correctly, email alerts will be disabled: %s", validationErr)
		}
	}
	APPLICATION_PROPERTIES, err = read_properties_file("application.properties")
	if err != nil {
		log.Printf("Failed to read application properties file: %s", err)
	}
	log.Printf("Application properties: %v", APPLICATION_PROPERTIES)

	DB, err = initialise_database(DATABASE_PROPERTIES)
	if err != nil {
		log.Fatalf("Failed to initialise database: %s", err)
	}

	err = create_temperature_readings_table()
	if err != nil {
		log.Fatalf("Failed to create temperature readings table: %s", err)
	}

	servers, err := discover_sensor_urls()
	if err != nil {
		log.Fatalf("Failed to discover sensor URLs: %s", err)
	}
	log.Printf("Identified the following servers with sensors: %s\n", servers)

	responses := take_readings(servers)

	err = add_list_of_readings(responses)
	if err != nil {
		log.Fatalf("Failed to add readings to the database: %s", err)
	}
	err = sendAlertEmailIfNeeded(responses)
	if err != nil {
		log.Printf("Failed to send alerts: %s", err)
	}
}
