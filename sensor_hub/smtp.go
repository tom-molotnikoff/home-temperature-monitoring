package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
)

func validateSMTPProperties() error {
	if SMTP_PROPERTIES["smtp.user"] == "" || SMTP_PROPERTIES["smtp.recipient"] == "" {
		return fmt.Errorf("smtp properties are not set correctly. Please check your smtp.properties file")
	}
	return nil
}

// This function sends an alert email using XOAUTH2 authentication.
// It uses the OAUTH_TOKEN which should be set after initialising OAuth2.
func sendAlertXOAUTH2(sensorName string, temperature float64) error {
	authStr := fmt.Sprintf("user=%s\001auth=Bearer %s\001\001", SMTP_PROPERTIES["smtp.user"], OAUTH_TOKEN.AccessToken)
	auth := smtp.Auth(&xoauth2Auth{SMTP_PROPERTIES["smtp.user"], OAUTH_TOKEN.AccessToken, authStr})

	subject := "Temperature Alert"
	body := "The temperature reading from sensor " + sensorName + " has breached a threshold, recorded temperature was: " +
		strconv.FormatFloat(temperature, 'f', 2, 64) +
		"°C"
	msg := "From: " + SMTP_PROPERTIES["smtp.user"] + "\n" +
		"To: " + SMTP_PROPERTIES["smtp.recipient"] + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	return smtp.SendMail("smtp.gmail.com:587", auth, SMTP_PROPERTIES["smtp.user"], []string{SMTP_PROPERTIES["smtp.recipient"]}, []byte(msg))
}

// This function checks if the temperature readings from sensors exceed the high or low thresholds
// defined in the application properties. If they do, it sends an alert email.
// It assumes that the temperature readings are in Celsius and that the thresholds are also defined in Celsius
func sendAlertEmailIfNeeded(responses []*SensorReading) error {
	if !OAUTH_SET {
		log.Println("OAuth2 is not set, skipping email alerts.")
		return nil
	}

	if APPLICATION_PROPERTIES["email.alert.high.threshold"] == "" {
		log.Println("No email alert threshold set, skipping email alerts.")
		return nil
	}
	highThreshold, err := strconv.ParseFloat(APPLICATION_PROPERTIES["email.alert.high.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid high threshold value: %s", APPLICATION_PROPERTIES["email.alert.high.threshold"])
	}
	if APPLICATION_PROPERTIES["email.alert.low.threshold"] == "" {
		log.Println("No low email alert threshold set, skipping email alerts.")
		return nil
	}
	lowThreshold, err := strconv.ParseFloat(APPLICATION_PROPERTIES["email.alert.low.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid low threshold value: %s", APPLICATION_PROPERTIES["email.alert.low.threshold"])
	}

	for _, reading := range responses {
		if reading.Reading.Temperature > highThreshold || reading.Reading.Temperature < lowThreshold { // Assuming 30.0°C is the threshold for alert
			err := sendAlertXOAUTH2(reading.SensorName, reading.Reading.Temperature)
			if err != nil {
				return fmt.Errorf("failed to send alert email for sensor %s: %s", reading.SensorName, err)
			} else {
				log.Printf("Alert email sent for sensor %s with temperature %.2f°C\n", reading.SensorName, reading.Reading.Temperature)
			}
		}
	}
	return nil
}
