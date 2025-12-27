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
	HealthHistoryRetentionDays         int
	SensorDataRetentionDays            int
	DataCleanupIntervalHours           int
	HealthHistoryDefaultResponseNumber int

	SMTPUser      string
	SMTPRecipient string

	DatabaseUsername string
	DatabasePassword string
	DatabaseHostname string
	DatabasePort     string
}

var AppConfig *ApplicationConfiguration

func SetHealthHistoryDefaultResponseNumber(number int) {
	AppConfig.HealthHistoryDefaultResponseNumber = number
}

func SetDataCleanupIntervalHours(hours int) {
	AppConfig.DataCleanupIntervalHours = hours
}

func SetHealthHistoryRetentionDays(days int) {
	AppConfig.HealthHistoryRetentionDays = days
}

func SetSensorDataRetentionDays(days int) {
	AppConfig.SensorDataRetentionDays = days
}

func SetEmailAlertHighTemperatureThreshold(threshold float64) {
	AppConfig.EmailAlertHighTemperatureThreshold = threshold
}

func SetEmailAlertLowTemperatureThreshold(threshold float64) {
	AppConfig.EmailAlertLowTemperatureThreshold = threshold
}

func SetSensorCollectionInterval(interval int) {
	AppConfig.SensorCollectionInterval = interval
}

func SetSensorDiscoverySkip(skip bool) {
	AppConfig.SensorDiscoverySkip = skip
}

func SetOpenAPILocation(location string) {
	AppConfig.OpenAPILocation = location
}

func SetSMTPUser(user string) {
	AppConfig.SMTPUser = user
}

func SetSMTPRecipient(recipient string) {
	AppConfig.SMTPRecipient = recipient
}

func SetDatabaseUsername(username string) {
	AppConfig.DatabaseUsername = username
}

func SetDatabasePassword(password string) {
	AppConfig.DatabasePassword = password
}

func SetDatabaseHostname(hostname string) {
	AppConfig.DatabaseHostname = hostname
}

func SetDatabasePort(port string) {
	AppConfig.DatabasePort = port
}

func ConvertConfigurationToMaps(cfg *ApplicationConfiguration) (map[string]string, map[string]string, map[string]string) {
	appProps := make(map[string]string)
	smtpProps := make(map[string]string)
	dbProps := make(map[string]string)

	appProps["email.alert.high.temperature.threshold"] = strconv.FormatFloat(cfg.EmailAlertHighTemperatureThreshold, 'f', -1, 64)
	appProps["email.alert.low.temperature.threshold"] = strconv.FormatFloat(cfg.EmailAlertLowTemperatureThreshold, 'f', -1, 64)
	appProps["sensor.collection.interval"] = strconv.Itoa(cfg.SensorCollectionInterval)
	appProps["sensor.discovery.skip"] = strconv.FormatBool(cfg.SensorDiscoverySkip)
	appProps["openapi.yaml.location"] = cfg.OpenAPILocation
	appProps["health.history.retention.days"] = strconv.Itoa(cfg.HealthHistoryRetentionDays)
	appProps["sensor.data.retention.days"] = strconv.Itoa(cfg.SensorDataRetentionDays)
	appProps["data.cleanup.interval.hours"] = strconv.Itoa(cfg.DataCleanupIntervalHours)
	appProps["health.history.default.response.number"] = strconv.Itoa(cfg.HealthHistoryDefaultResponseNumber)

	smtpProps["smtp.user"] = cfg.SMTPUser
	smtpProps["smtp.recipient"] = cfg.SMTPRecipient

	dbProps["database.username"] = cfg.DatabaseUsername
	dbProps["database.password"] = cfg.DatabasePassword
	dbProps["database.hostname"] = cfg.DatabaseHostname
	dbProps["database.port"] = cfg.DatabasePort

	return appProps, smtpProps, dbProps
}

func LoadConfigurationFromMaps(appProps, smtpProps, dbProps map[string]string) (*ApplicationConfiguration, error) {
	cfg := &ApplicationConfiguration{}

	if v, ok := appProps["email.alert.high.temperature.threshold"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.EmailAlertHighTemperatureThreshold = f
		} else {
			log.Printf("invalid high temp threshold '%s': %v", v, err)
			return nil, err
		}
	}
	if v, ok := appProps["email.alert.low.temperature.threshold"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			cfg.EmailAlertLowTemperatureThreshold = f
		} else {
			log.Printf("invalid low temp threshold '%s': %v", v, err)
			return nil, err
		}
	}

	if v, ok := appProps["sensor.collection.interval"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SensorCollectionInterval = i
		} else {
			log.Printf("invalid sensor.collection.interval '%s': %v", v, err)
			return nil, err
		}
	}

	cfg.OpenAPILocation = appProps["openapi.yaml.location"]
	if v, ok := appProps["sensor.discovery.skip"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.SensorDiscoverySkip = b
		} else {
			log.Printf("invalid sensor.discovery.skip '%s': %v", v, err)
			return nil, err
		}
	}

	if v, ok := appProps["health.history.retention.days"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.HealthHistoryRetentionDays = i
		} else {
			log.Printf("invalid health.history.retention.days '%s': %v", v, err)
			return nil, err
		}
	}

	if v, ok := appProps["sensor.data.retention.days"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SensorDataRetentionDays = i
		} else {
			log.Printf("invalid sensor.data.retention.days '%s': %v", v, err)
			return nil, err
		}
	}

	if v, ok := appProps["data.cleanup.interval.hours"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.DataCleanupIntervalHours = i
		} else {
			log.Printf("invalid data.cleanup.interval.hours '%s': %v", v, err)
			return nil, err
		}
	}

	if v, ok := appProps["health.history.default.response.number"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.HealthHistoryDefaultResponseNumber = i
		} else {
			log.Printf("invalid health.history.default.response.number '%s': %v", v, err)
			return nil, err
		}
	}

	cfg.SMTPUser = smtpProps["smtp.user"]
	cfg.SMTPRecipient = smtpProps["smtp.recipient"]

	cfg.DatabaseUsername = dbProps["database.username"]
	cfg.DatabasePassword = dbProps["database.password"]
	cfg.DatabaseHostname = dbProps["database.hostname"]
	cfg.DatabasePort = dbProps["database.port"]

	return cfg, nil
}

func InitialiseConfig() error {
	appProps, err := ReadApplicationPropertiesFile()
	if err != nil {
		return err
	}

	smtpProps, err := ReadSMTPPropertiesFile()
	if err != nil {
		return err
	}

	dbProps, err := ReadDatabasePropertiesFile()
	if err != nil {
		return err
	}

	ReloadConfig(appProps, smtpProps, dbProps)

	return nil
}

func ReloadConfig(appProps, smtpProps, dbProps map[string]string) {
	cfg, err := LoadConfigurationFromMaps(appProps, smtpProps, dbProps)
	if err != nil {
		log.Printf("Failed to reload configuration: %v", err)
		return
	}

	AppConfig = cfg

	// Don't include the sensitive ones!
	log.Printf("Configuration reloaded: %+v", struct {
		EmailAlertHighTemperatureThreshold float64
		EmailAlertLowTemperatureThreshold  float64
		SensorCollectionInterval           int
		SensorDiscoverySkip                bool
		OpenAPILocation                    string
		HealthHistoryRetentionDays         int
		SensorDataRetentionDays            int
		DataCleanupIntervalHours           int
		HealthHistoryDefaultResponseNumber int
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
		AppConfig.HealthHistoryRetentionDays,
		AppConfig.SensorDataRetentionDays,
		AppConfig.DataCleanupIntervalHours,
		AppConfig.HealthHistoryDefaultResponseNumber,
		AppConfig.SMTPUser,
		AppConfig.SMTPRecipient,
		AppConfig.DatabaseUsername,
		AppConfig.DatabaseHostname,
		AppConfig.DatabasePort,
	})
}
