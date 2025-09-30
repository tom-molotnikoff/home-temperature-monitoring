package smtp

import (
	appProps "example/sensorHub/application_properties"
	"example/sensorHub/oauth"
	"example/sensorHub/types"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
)

// This function sends an alert email using XOAUTH2 authentication.
// It uses the OAUTH_TOKEN which should be set after initialising OAuth2.
func SendAlertXOAUTH2(sensorName string, temperature float64) error {
	authStr := fmt.Sprintf("user=%s\001auth=Bearer %s\001\001", appProps.SMTP_PROPERTIES["smtp.user"], oauth.OAUTH_TOKEN.AccessToken)
	auth := smtp.Auth(&types.XOauth2Auth{
		Username:    appProps.SMTP_PROPERTIES["smtp.user"],
		AccessToken: oauth.OAUTH_TOKEN.AccessToken,
		AuthString:  authStr,
	})

	subject := "Temperature Alert"
	body := "The temperature reading from sensor " + sensorName + " has breached a threshold, recorded temperature was: " +
		strconv.FormatFloat(temperature, 'f', 2, 64) +
		"°C"
	msg := "From: " + appProps.SMTP_PROPERTIES["smtp.user"] + "\n" +
		"To: " + appProps.SMTP_PROPERTIES["smtp.recipient"] + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	return smtp.SendMail("smtp.gmail.com:587", auth, appProps.SMTP_PROPERTIES["smtp.user"], []string{appProps.SMTP_PROPERTIES["smtp.recipient"]}, []byte(msg))
}

// This function checks if the temperature readings from sensors exceed the high or low thresholds
// defined in the application properties. If they do, it sends an alert email.
// It assumes that the temperature readings are in Celsius and that the thresholds are also defined in Celsius
func SendAlertEmailIfNeeded(responses []types.APIReading) error {
	if !oauth.OAUTH_SET {
		log.Println("OAuth2 is not set, skipping email alerts.")
		return nil
	}

	if appProps.APPLICATION_PROPERTIES["email.alert.high.threshold"] == "" {
		log.Println("No email alert threshold set, skipping email alerts.")
		return nil
	}
	highThreshold, err := strconv.ParseFloat(appProps.APPLICATION_PROPERTIES["email.alert.high.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid high threshold value: %s", appProps.APPLICATION_PROPERTIES["email.alert.high.threshold"])
	}
	if appProps.APPLICATION_PROPERTIES["email.alert.low.threshold"] == "" {
		log.Println("No low email alert threshold set, skipping email alerts.")
		return nil
	}
	lowThreshold, err := strconv.ParseFloat(appProps.APPLICATION_PROPERTIES["email.alert.low.threshold"], 64)
	if err != nil {
		return fmt.Errorf("invalid low threshold value: %s", appProps.APPLICATION_PROPERTIES["email.alert.low.threshold"])
	}

	for _, reading := range responses {
		if reading.Reading.Temperature > highThreshold || reading.Reading.Temperature < lowThreshold { // Assuming 30.0°C is the threshold for alert
			err := SendAlertXOAUTH2(reading.SensorName, reading.Reading.Temperature)
			if err != nil {
				return fmt.Errorf("failed to send alert email for sensor %s: %w", reading.SensorName, err)
			} else {
				log.Printf("Alert email sent for sensor %s with temperature %.2f°C\n", reading.SensorName, reading.Reading.Temperature)
			}
		}
	}
	return nil
}
