package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GET /sensors/temperature
// This handler will collect temperature readings from all sensors
// and return them as a JSON response.
func collect_all_sensors_handler(ctx *gin.Context) {
	log.Println("Collecting all sensor readings...")
	readings, err := take_readings()
	if err != nil {
		log.Printf("Error collecting readings: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error collecting readings"})
		return
	}

	ctx.IndentedJSON(http.StatusOK, readings)
}

// GET /sensors/temperature/:sensorName
// This handler will retrieve the temperature reading for a specific sensor
// based on the sensor name provided in the URL.
// It will return the reading as a JSON response.
func collect_specific_sensor_handler(ctx *gin.Context) {
	sensorName := ctx.Param("sensorName")
	log.Printf("Retrieving sensor reading for sensor: %s", sensorName)
	reading, err := take_reading_from_named_sensor(sensorName)

	if err != nil {
		log.Printf("Error retrieving reading for sensor %s: %s", sensorName, err)
		ctx.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, reading)
}

// GET /readings/between
// This handler will retrieve temperature readings between two dates.
// It will parse the start and end dates from the query parameters,
// fetch the readings from the database, and return them as a JSON array.
func get_readings_between_dates_handler(ctx *gin.Context) {
	startDate := ctx.Query("start")
	endDate := ctx.Query("end")
	if startDate == "" || endDate == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Start and end dates are required"})
		return
	}

	_, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid start date format, expected YYYY-MM-DD"})
		return
	}

	_, err = time.Parse("2006-01-02", endDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid end date format, expected YYYY-MM-DD"})
		return
	}

	log.Printf("Fetching readings between %s and %s", startDate, endDate)
	readings, err := getReadingsBetweenDates(startDate, endDate)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	ctx.IndentedJSON(http.StatusOK, readings)
}

// currentTemperaturesWebSocket handles WebSocket connections to provide
// real-time temperature readings from all sensors.
func currentTemperaturesWebSocket(c *gin.Context) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	interval := APPLICATION_PROPERTIES["current.temperature.websocket.interval"]
	if interval == "" {
		interval = "10" // Default to 5 seconds if not set
	}
	intervalDuration, err := time.ParseDuration(interval + "s")
	if err != nil {
		log.Printf("Invalid interval duration: %v, using default 10 seconds", err)
		intervalDuration = 10 * time.Second // Default to 10 seconds
	}
	ticker := time.NewTicker(intervalDuration)
	defer ticker.Stop()

	// Channel to signal close
	done := make(chan struct{})

	// Goroutine to listen for client close
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				close(done)
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			readings, err := getLatestReadings()
			if err != nil {
				log.Printf("Error fetching latest readings: %v", err)
				continue
			}
			if err := conn.WriteJSON(readings); err != nil {
				log.Printf("WebSocket closed or error: %v", err)
				return // Exit the handler when the connection is closed
			}
		case <-done:
			log.Printf("WebSocket connection closed by client")
			return
		}
	}
}

// This function will set up the API server and start listening for requests.
// It will use the discovered sensor URLs to fetch temperature readings and
// handle incoming requests to retrieve these readings.
func initalise_api_and_listen() {
	log.Println("API server is starting...")
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // For development, allow all
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/sensors/temperature", collect_all_sensors_handler)
	router.GET("/sensors/temperature/:sensorName", collect_specific_sensor_handler)
	router.GET("/readings/between", get_readings_between_dates_handler)
	router.GET("/ws/current-temperatures", currentTemperaturesWebSocket)
	log.Println("API server is running on port 8080")
	router.Run("0.0.0.0:8080")
}
