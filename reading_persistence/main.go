package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Servers struct {
	Servers []ServerItem `yaml:"servers"`
}

type ServerItem struct {
	Url string `yaml:"url"`
}

func discover_sensor_urls() []string {
	fileData, err := os.ReadFile("../openapi.yaml")
	if err != nil {
		log.Fatalf("Cannot find the openapi.yaml file for the temperature sensors: %s\n", err)
	}
	var servers Servers

	err = yaml.Unmarshal(fileData, &servers)
	if err != nil {
		log.Fatalf("Cannot unmarshal the yaml into a map: %s\n", err)
	}
	urls := make([]string, 0)

	for _, value := range servers.Servers {
		urls = append(urls, value.Url)
	}
	return urls
}

type SensorReading struct {
	Name    string
	Reading TemperatureReading
}

type TemperatureReading struct {
	Temperature float64
	Time        string
}

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

func main() {
	servers := discover_sensor_urls()
	log.Printf("Identified the following servers with sensors: %s\n", servers)
	responses := take_readings(servers)
	db_properties := read_db_properties_file("database.properties")
	DB = initialise_database(db_properties)
	create_temperature_readings_table()
	add_list_of_readings(responses)
}
