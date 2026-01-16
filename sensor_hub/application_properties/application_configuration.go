package appProps

import (
	"log"
	"strconv"
)

type ApplicationConfiguration struct {
	// DEPRECATED: Alert thresholds are now stored per-sensor in the database (sensor_alert_rules table).
	// These fields are retained for backwards compatibility with existing application.properties files
	// but are no longer used by the alerting system. See V14 migration for database-driven alerting.
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

	AuthBcryptCost                int
	AuthSessionTTLMinutes         int
	AuthSessionCookieName         string
	AuthLoginBackoffWindowMinutes int
	AuthLoginBackoffThreshold     int
	AuthLoginBackoffBaseSeconds   int
	AuthLoginBackoffMaxSeconds    int

	FailedLoginRetentionDays int
}

var AppConfig *ApplicationConfiguration

func SetHealthHistoryDefaultResponseNumber(number int) {
	AppConfig.HealthHistoryDefaultResponseNumber = number
}

func SetFailedLoginRetentionDays(days int) {
	AppConfig.FailedLoginRetentionDays = days
}

func SetAuthBcryptCost(cost int) {
	AppConfig.AuthBcryptCost = cost
}

func SetAuthSessionTTLMinutes(ttl int) {
	AppConfig.AuthSessionTTLMinutes = ttl
}

func SetAuthSessionCookieName(name string) {
	AppConfig.AuthSessionCookieName = name
}

func SetAuthLoginBackoffWindowMinutes(minutes int) {
	AppConfig.AuthLoginBackoffWindowMinutes = minutes
}

func SetAuthLoginBackoffThreshold(threshold int) {
	AppConfig.AuthLoginBackoffThreshold = threshold
}

func SetAuthLoginBackoffBaseSeconds(seconds int) {
	AppConfig.AuthLoginBackoffBaseSeconds = seconds
}

func SetAuthLoginBackoffMaxSeconds(seconds int) {
	AppConfig.AuthLoginBackoffMaxSeconds = seconds
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

// DEPRECATED: Alert thresholds are now stored per-sensor in the database.
// This setter is retained for backwards compatibility but has no effect on alerting behavior.
func SetEmailAlertHighTemperatureThreshold(threshold float64) {
	AppConfig.EmailAlertHighTemperatureThreshold = threshold
}

// DEPRECATED: Alert thresholds are now stored per-sensor in the database.
// This setter is retained for backwards compatibility but has no effect on alerting behavior.
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
	appProps["failed.login.retention.days"] = strconv.Itoa(cfg.FailedLoginRetentionDays)

	// auth
	appProps["auth.bcrypt.cost"] = strconv.Itoa(cfg.AuthBcryptCost)
	appProps["auth.session.ttl.minutes"] = strconv.Itoa(cfg.AuthSessionTTLMinutes)
	appProps["auth.session.cookie.name"] = cfg.AuthSessionCookieName
	appProps["auth.login.backoff.window.minutes"] = strconv.Itoa(cfg.AuthLoginBackoffWindowMinutes)
	appProps["auth.login.backoff.threshold"] = strconv.Itoa(cfg.AuthLoginBackoffThreshold)
	appProps["auth.login.backoff.base.seconds"] = strconv.Itoa(cfg.AuthLoginBackoffBaseSeconds)
	appProps["auth.login.backoff.max.seconds"] = strconv.Itoa(cfg.AuthLoginBackoffMaxSeconds)

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

	if v, ok := appProps["failed.login.retention.days"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.FailedLoginRetentionDays = i
		} else {
			log.Printf("invalid failed.login.retention.days '%s': %v", v, err)
			return nil, err
		}
	}

	// auth
	if v, ok := appProps["auth.bcrypt.cost"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthBcryptCost = i
		} else {
			log.Printf("invalid auth.bcrypt.cost '%s': %v", v, err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.session.ttl.minutes"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthSessionTTLMinutes = i
		} else {
			log.Printf("invalid auth.session.ttl.minutes '%s': %v", v, err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.session.cookie.name"]; ok {
		cfg.AuthSessionCookieName = v
	}
	if v, ok := appProps["auth.login.backoff.window.minutes"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffWindowMinutes = i
		} else {
			log.Printf("invalid auth.login.backoff.window.minutes '%s': %v", v, err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.login.backoff.threshold"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffThreshold = i
		} else {
			log.Printf("invalid auth.login.backoff.threshold '%s': %v", v, err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.login.backoff.base.seconds"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffBaseSeconds = i
		} else {
			log.Printf("invalid auth.login.backoff.base.seconds '%s': %v", v, err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.login.backoff.max.seconds"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffMaxSeconds = i
		} else {
			log.Printf("invalid auth.login.backoff.max.seconds '%s': %v", v, err)
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
		FailedLoginRetentionDays           int
		AuthBcryptCost                     int
		AuthSessionTTLMinutes              int
		AuthSessionCookieName              string
		AuthLoginBackoffWindowMinutes      int
		AuthLoginBackoffThreshold          int
		AuthLoginBackoffBaseSeconds        int
		AuthLoginBackoffMaxSeconds         int
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
		AppConfig.FailedLoginRetentionDays,
		AppConfig.AuthBcryptCost,
		AppConfig.AuthSessionTTLMinutes,
		AppConfig.AuthSessionCookieName,
		AppConfig.AuthLoginBackoffWindowMinutes,
		AppConfig.AuthLoginBackoffThreshold,
		AppConfig.AuthLoginBackoffBaseSeconds,
		AppConfig.AuthLoginBackoffMaxSeconds,
		AppConfig.SMTPUser,
		AppConfig.SMTPRecipient,
		AppConfig.DatabaseUsername,
		AppConfig.DatabaseHostname,
		AppConfig.DatabasePort,
	})
}
