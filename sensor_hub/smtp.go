package main

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
)

func validateSMTPProperties() error {
	if SMTP_PROPERTIES["smtp.user"] == "" || SMTP_PROPERTIES["smtp.pass"] == "" ||
		SMTP_PROPERTIES["smtp.host"] == "" || SMTP_PROPERTIES["smtp.port"] == "" ||
		SMTP_PROPERTIES["smtp.recipient"] == "" {
		return fmt.Errorf("smtp properties are not set correctly. Please check your smtp.properties file")
	}
	return nil
}

// This function will send an alert email if the temperature exceeds a threshold
// Requires smtpUser, smtpPass, smtpHost, smtpPort, and smtpRecipient with your SMTP server details
// to be provided through a properties file.
func sendAlertEmail(sensorName string, temperature float64) error {
	log.Printf("Sending alert email for sensor %s with temperature %.2f\n", sensorName, temperature)

	subject := "Temperature Alert"
	body := "The temperature reading from sensor " + sensorName + " has breached a threshold, recorded temperature was: " +
		strconv.FormatFloat(temperature, 'f', 2, 64) +
		"°C"
	msg := "From: " + SMTP_PROPERTIES["smtp.user"] + "\n" +
		"To: " + SMTP_PROPERTIES["smtp.recipient"] + "\n" +
		"Subject: " + subject + "\n\n" +
		body
	auth := smtp.PlainAuth("", SMTP_PROPERTIES["smtp.user"], SMTP_PROPERTIES["smtp.pass"], SMTP_PROPERTIES["smtp.host"])
	return smtp.SendMail(SMTP_PROPERTIES["smtp.host"]+":"+SMTP_PROPERTIES["smtp.port"], auth, SMTP_PROPERTIES["smtp.user"], []string{SMTP_PROPERTIES["smtp.recipient"]}, []byte(msg))
}

// This function checks if the temperature readings from sensors exceed the high or low thresholds
// defined in the application properties. If they do, it sends an alert email.
// It assumes that the temperature readings are in Celsius and that the thresholds are also defined in Celsius
func sendAlertEmailIfNeeded(responses []*SensorReading) error {
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
			err := sendAlertEmail(reading.Name, reading.Reading.Temperature)
			if err != nil {
				return fmt.Errorf("failed to send alert email for sensor %s: %s", reading.Name, err)
			} else {
				log.Printf("Alert email sent for sensor %s with temperature %.2f°C\n", reading.Name, reading.Reading.Temperature)
			}
		}
	}
	return nil
}
