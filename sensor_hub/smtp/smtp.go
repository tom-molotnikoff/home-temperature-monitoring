package smtp

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"

	appProps "example/sensorHub/application_properties"
	"example/sensorHub/oauth"
)

type SMTPNotifier struct{}

func NewSMTPNotifier() *SMTPNotifier {
	return &SMTPNotifier{}
}

func (n *SMTPNotifier) SendAlert(sensorName, sensorType, reason string, numericValue float64, statusValue string) error {
	if !oauth.OauthSet {
		log.Printf("OAuth not configured, skipping alert for sensor %s", sensorName)
		return nil
	}

	subject, body := formatAlertMessage(sensorName, sensorType, reason, numericValue, statusValue)

	auth := &oauth.XOauth2Auth{
		Username:    appProps.AppConfig.SMTPUser,
		AccessToken: oauth.OauthToken.AccessToken,
	}

	msg := "From: " + appProps.AppConfig.SMTPUser + "\n" +
		"To: " + appProps.AppConfig.SMTPRecipient + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail("smtp.gmail.com:587", auth, appProps.AppConfig.SMTPUser, []string{appProps.AppConfig.SMTPRecipient}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

func formatAlertMessage(sensorName, sensorType, reason string, numericValue float64, statusValue string) (string, string) {
	subject := fmt.Sprintf("Sensor Alert: %s", sensorName)

	var body string
	if statusValue != "" {
		body = fmt.Sprintf(
			"Alert from sensor: %s\nType: %s\nStatus: %s\n\nReason: %s",
			sensorName,
			sensorType,
			statusValue,
			reason,
		)
	} else {
		body = fmt.Sprintf(
			"Alert from sensor: %s\nType: %s\nValue: %s\n\nReason: %s",
			sensorName,
			sensorType,
			strconv.FormatFloat(numericValue, 'f', 2, 64),
			reason,
		)
	}

	return subject, body
}
