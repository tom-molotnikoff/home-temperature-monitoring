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

func SendAlertXOAUTH2(sensorName string, temperature float64) error {
	authStr := fmt.Sprintf("user=%s\001auth=Bearer %s\001\001", appProps.AppConfig.SMTPUser, oauth.OauthToken.AccessToken)
	auth := smtp.Auth(&oauth.XOauth2Auth{
		Username:    appProps.AppConfig.SMTPUser,
		AccessToken: oauth.OauthToken.AccessToken,
		AuthString:  authStr,
	})

	subject := "Temperature Alert"
	body := "The temperature reading from sensor " + sensorName + " has breached a threshold, recorded temperature was: " +
		strconv.FormatFloat(temperature, 'f', 2, 64) +
		"°C"
	msg := "From: " + appProps.AppConfig.SMTPUser + "\n" +
		"To: " + appProps.AppConfig.SMTPRecipient + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	return smtp.SendMail("smtp.gmail.com:587", auth, appProps.AppConfig.SMTPUser, []string{appProps.AppConfig.SMTPRecipient}, []byte(msg))
}

func SendTemperatureAlertEmailIfNeeded(responses []types.TemperatureReading) error {
	if !oauth.OauthSet {
		return nil
	}

	highThreshold := appProps.AppConfig.EmailAlertHighTemperatureThreshold
	lowThreshold := appProps.AppConfig.EmailAlertLowTemperatureThreshold

	for _, reading := range responses {
		if reading.Temperature > highThreshold || reading.Temperature < lowThreshold {
			err := SendAlertXOAUTH2(reading.SensorName, reading.Temperature)
			if err != nil {
				return fmt.Errorf("failed to send alert email for sensor %s: %w", reading.SensorName, err)
			}
			log.Printf("Alert email sent for sensor %s with temperature %.2f°C\n", reading.SensorName, reading.Temperature)
		}
	}
	return nil
}
