package appProps

import (
	"log"
	"strconv"
)

type ApplicationConfiguration struct {
	EmailAlertHighTemperatureThreshold float64
	EmailAlertLowTemperatureThreshold  float64
	SensorCollectionInterval           int
	SensorDiscoverySkip                bool
	OpenAPILocation                    string

	SMTPUser      string
	SMTPRecipient string

	DatabaseUsername string
	DatabasePassword string
	DatabaseHostname string
	DatabasePort     string
}

var AppConfig *ApplicationConfiguration

func LoadConfigurationFromMaps(appProps, smtpProps, dbProps map[string]string) *ApplicationConfiguration {
	cfg := &ApplicationConfiguration{}

	// parse floats
	if v, ok := appProps["email.alert.high.temperature.threshold"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.EmailAlertHighTemperatureThreshold = f
		} else {
			log.Printf("invalid high temp threshold '%s': %v", v, err)
		}
	}
	if v, ok := appProps["email.alert.low.temperature.threshold"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.EmailAlertLowTemperatureThreshold = f
		} else {
			log.Printf("invalid low temp threshold '%s': %v", v, err)
		}
	}

	// parse int
	if v, ok := appProps["sensor.collection.interval"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SensorCollectionInterval = i
		} else {
			log.Printf("invalid sensor.collection.interval '%s': %v", v, err)
		}
	}

	// simple strings / bool
	cfg.OpenAPILocation = appProps["openapi.yaml.location"]
	if v, ok := appProps["sensor.discovery.skip"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.SensorDiscoverySkip = b
		} else {
			log.Printf("invalid sensor.discovery.skip '%s': %v", v, err)
		}
	}

	// smtp props
	cfg.SMTPUser = smtpProps["smtp.user"]
	cfg.SMTPRecipient = smtpProps["smtp.recipient"]

	// database props
	cfg.DatabaseUsername = dbProps["database.username"]
	cfg.DatabasePassword = dbProps["database.password"]
	cfg.DatabaseHostname = dbProps["database.hostname"]
	cfg.DatabasePort = dbProps["database.port"]

	return cfg
}

func ReloadConfig() {
	AppConfig = LoadConfigurationFromMaps(applicationProperties, smtpProperties, databaseProperties)

	// Don't include the sensitive ones!
	log.Printf("Configuration reloaded: %+v", struct {
		EmailAlertHighTemperatureThreshold float64
		EmailAlertLowTemperatureThreshold  float64
		SensorCollectionInterval           int
		SensorDiscoverySkip                bool
		OpenAPILocation                    string
		SMTPUser                           string
		SMTPRecipient                      string
		DatabaseUsername                   string
		DatabaseHostname                   string
		DatabasePort                       string
	}{
		AppConfig.EmailAlertHighTemperatureThreshold,
		AppConfig.EmailAlertLowTemperatureThreshold,
		AppConfig.SensorCollectionInterval,
		AppConfig.SensorDiscoverySkip,
		AppConfig.OpenAPILocation,
		AppConfig.SMTPUser,
		AppConfig.SMTPRecipient,
		AppConfig.DatabaseUsername,
		AppConfig.DatabaseHostname,
		AppConfig.DatabasePort,
	})
}
