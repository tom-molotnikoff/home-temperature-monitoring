package smtp

import (
	"fmt"
	"log"
	"net/smtp"

	appProps "example/sensorHub/application_properties"
	"example/sensorHub/oauth"
)

type SMTPNotifier struct{}

func NewSMTPNotifier() *SMTPNotifier {
	return &SMTPNotifier{}
}

func (n *SMTPNotifier) SendNotification(recipient, title, message, category string) error {
	if !oauth.OauthSet {
		log.Printf("OAuth not configured, skipping notification email to %s", recipient)
		return nil
	}

	// Get current token from OAuth service (may have been refreshed)
	token := oauth.OauthToken
	if token == nil {
		log.Printf("OAuth token is nil, skipping notification email to %s", recipient)
		return nil
	}

	auth := &oauth.XOauth2Auth{
		Username:    appProps.AppConfig.SMTPUser,
		AccessToken: token.AccessToken,
	}

	subject := fmt.Sprintf("[%s] %s", category, title)
	msg := "From: " + appProps.AppConfig.SMTPUser + "\n" +
		"To: " + recipient + "\n" +
		"Subject: " + subject + "\n\n" +
		message

	err := smtp.SendMail("smtp.gmail.com:587", auth, appProps.AppConfig.SMTPUser, []string{recipient}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send notification email via SMTP: %w", err)
	}

	log.Printf("Notification email sent to %s: %s", recipient, title)
	return nil
}
