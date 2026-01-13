package smtp

import (
	"testing"

	appProps "example/sensorHub/application_properties"
	"example/sensorHub/oauth"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestNewSMTPNotifier(t *testing.T) {
	notifier := NewSMTPNotifier()
	assert.NotNil(t, notifier)
}

func TestSMTPNotifier_SendAlert_OAuthNotSet(t *testing.T) {
	originalOauthSet := oauth.OauthSet
	defer func() { oauth.OauthSet = originalOauthSet }()

	oauth.OauthSet = false

	notifier := NewSMTPNotifier()
	err := notifier.SendAlert("TestSensor", "temperature", "above threshold", 35.0, "")

	assert.NoError(t, err)
}

func TestSMTPNotifier_SendAlert_NumericSensor(t *testing.T) {
	originalOauthSet := oauth.OauthSet
	originalToken := oauth.OauthToken
	originalConfig := appProps.AppConfig
	defer func() {
		oauth.OauthSet = originalOauthSet
		oauth.OauthToken = originalToken
		appProps.AppConfig = originalConfig
	}()

	oauth.OauthSet = true
	oauth.OauthToken = &oauth2.Token{
		AccessToken: "test-token",
	}
	appProps.AppConfig = &appProps.ApplicationConfiguration{
		SMTPUser:      "test@example.com",
		SMTPRecipient: "recipient@example.com",
	}

	notifier := NewSMTPNotifier()
	err := notifier.SendAlert("TestSensor", "temperature", "value 35.00 is above high threshold 30.00", 35.0, "")

	assert.Error(t, err)
}

func TestSMTPNotifier_SendAlert_StatusSensor(t *testing.T) {
	originalOauthSet := oauth.OauthSet
	originalToken := oauth.OauthToken
	originalConfig := appProps.AppConfig
	defer func() {
		oauth.OauthSet = originalOauthSet
		oauth.OauthToken = originalToken
		appProps.AppConfig = originalConfig
	}()

	oauth.OauthSet = true
	oauth.OauthToken = &oauth2.Token{
		AccessToken: "test-token",
	}
	appProps.AppConfig = &appProps.ApplicationConfiguration{
		SMTPUser:      "test@example.com",
		SMTPRecipient: "recipient@example.com",
	}

	notifier := NewSMTPNotifier()
	err := notifier.SendAlert("BackDoor", "door", "status is open", 0, "open")

	assert.Error(t, err)
}

func TestFormatAlertMessage_NumericSensor(t *testing.T) {
	subject, body := formatAlertMessage("KitchenTemp", "temperature", "value 35.50 is above high threshold 30.00", 35.5, "")

	assert.Equal(t, "Sensor Alert: KitchenTemp", subject)
	assert.Contains(t, body, "KitchenTemp")
	assert.Contains(t, body, "temperature")
	assert.Contains(t, body, "35.50")
	assert.Contains(t, body, "value 35.50 is above high threshold 30.00")
}

func TestFormatAlertMessage_StatusSensor(t *testing.T) {
	subject, body := formatAlertMessage("BackDoor", "door", "status is open", 0, "open")

	assert.Equal(t, "Sensor Alert: BackDoor", subject)
	assert.Contains(t, body, "BackDoor")
	assert.Contains(t, body, "door")
	assert.Contains(t, body, "open")
	assert.Contains(t, body, "status is open")
	assert.NotContains(t, body, "0.00")
}
