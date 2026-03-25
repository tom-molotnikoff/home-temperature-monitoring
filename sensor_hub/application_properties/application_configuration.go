package appProps

import (
	"log"
	"path/filepath"
	"strconv"
)

type ApplicationConfiguration struct {
	SensorCollectionInterval           int
	SensorDiscoverySkip                bool
	OpenAPILocation                    string
	HealthHistoryRetentionDays         int
	SensorDataRetentionDays            int
	FailedLoginRetentionDays           int
	DataCleanupIntervalHours           int
	HealthHistoryDefaultResponseNumber int

	SMTPUser string

	DatabasePath string

	AuthBcryptCost                int
	AuthSessionTTLMinutes         int
	AuthSessionCookieName         string
	AuthLoginBackoffWindowMinutes int
	AuthLoginBackoffThreshold     int
	AuthLoginBackoffBaseSeconds   int
	AuthLoginBackoffMaxSeconds    int

	// OAuth configuration
	OAuthCredentialsFilePath         string
	OAuthTokenFilePath               string
	OAuthTokenRefreshIntervalMinutes int
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

func SetDatabasePath(path string) {
	AppConfig.DatabasePath = path
}

func SetOAuthCredentialsFilePath(path string) {
	AppConfig.OAuthCredentialsFilePath = path
}

func SetOAuthTokenFilePath(path string) {
	AppConfig.OAuthTokenFilePath = path
}

func SetOAuthTokenRefreshIntervalMinutes(minutes int) {
	AppConfig.OAuthTokenRefreshIntervalMinutes = minutes
}

func ConvertConfigurationToMaps(cfg *ApplicationConfiguration) (map[string]string, map[string]string, map[string]string) {
	appProps := make(map[string]string)
	smtpProps := make(map[string]string)
	dbProps := make(map[string]string)

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

	// OAuth
	appProps["oauth.credentials.file.path"] = cfg.OAuthCredentialsFilePath
	appProps["oauth.token.file.path"] = cfg.OAuthTokenFilePath
	appProps["oauth.token.refresh.interval.minutes"] = strconv.Itoa(cfg.OAuthTokenRefreshIntervalMinutes)

	smtpProps["smtp.user"] = cfg.SMTPUser

	dbProps["database.path"] = cfg.DatabasePath

	return appProps, smtpProps, dbProps
}

func LoadConfigurationFromMaps(appProps, smtpProps, dbProps map[string]string) (*ApplicationConfiguration, error) {
	cfg := &ApplicationConfiguration{}

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

	// OAuth
	if v, ok := appProps["oauth.credentials.file.path"]; ok {
		if !filepath.IsAbs(v) {
			v = filepath.Join(configDir, v)
		}
		cfg.OAuthCredentialsFilePath = v
	}
	if v, ok := appProps["oauth.token.file.path"]; ok {
		if !filepath.IsAbs(v) {
			v = filepath.Join(configDir, v)
		}
		cfg.OAuthTokenFilePath = v
	}
	if v, ok := appProps["oauth.token.refresh.interval.minutes"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.OAuthTokenRefreshIntervalMinutes = i
		} else {
			log.Printf("invalid oauth.token.refresh.interval.minutes '%s': %v", v, err)
			return nil, err
		}
	}

	cfg.SMTPUser = smtpProps["smtp.user"]

	cfg.DatabasePath = dbProps["database.path"]

	return cfg, nil
}

func InitialiseConfig(dir string) error {
	setConfigPaths(dir)

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
		OAuthCredentialsFilePath           string
		OAuthTokenFilePath                 string
		OAuthTokenRefreshIntervalMinutes   int
		SMTPUser                           string
		DatabasePath                       string
	}{
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
		AppConfig.OAuthCredentialsFilePath,
		AppConfig.OAuthTokenFilePath,
		AppConfig.OAuthTokenRefreshIntervalMinutes,
		AppConfig.SMTPUser,
		AppConfig.DatabasePath,
	})
}
