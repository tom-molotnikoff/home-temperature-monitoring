package appProps

import (
	"log/slog"
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

	// Weather configuration
	WeatherLatitude     string
	WeatherLongitude    string
	WeatherLocationName string

	// Telemetry
	LogLevel string
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

	// Weather
	appProps["weather.latitude"] = cfg.WeatherLatitude
	appProps["weather.longitude"] = cfg.WeatherLongitude
	appProps["weather.location.name"] = cfg.WeatherLocationName

	// Telemetry
	appProps["log.level"] = cfg.LogLevel

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
			slog.Error("invalid config value", "key", "sensor.collection.interval", "value", v, "error", err)
			return nil, err
		}
	}

	cfg.OpenAPILocation = appProps["openapi.yaml.location"]
	if v, ok := appProps["sensor.discovery.skip"]; ok {
		if b, err := strconv.ParseBool(v); err == nil {
			cfg.SensorDiscoverySkip = b
		} else {
			slog.Error("invalid config value", "key", "sensor.discovery.skip", "value", v, "error", err)
			return nil, err
		}
	}

	if v, ok := appProps["health.history.retention.days"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.HealthHistoryRetentionDays = i
		} else {
			slog.Error("invalid config value", "key", "health.history.retention.days", "value", v, "error", err)
			return nil, err
		}
	}

	if v, ok := appProps["sensor.data.retention.days"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.SensorDataRetentionDays = i
		} else {
			slog.Error("invalid config value", "key", "sensor.data.retention.days", "value", v, "error", err)
			return nil, err
		}
	}

	if v, ok := appProps["data.cleanup.interval.hours"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.DataCleanupIntervalHours = i
		} else {
			slog.Error("invalid config value", "key", "data.cleanup.interval.hours", "value", v, "error", err)
			return nil, err
		}
	}

	if v, ok := appProps["health.history.default.response.number"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.HealthHistoryDefaultResponseNumber = i
		} else {
			slog.Error("invalid config value", "key", "health.history.default.response.number", "value", v, "error", err)
			return nil, err
		}
	}

	if v, ok := appProps["failed.login.retention.days"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.FailedLoginRetentionDays = i
		} else {
			slog.Error("invalid config value", "key", "failed.login.retention.days", "value", v, "error", err)
			return nil, err
		}
	}

	// auth
	if v, ok := appProps["auth.bcrypt.cost"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthBcryptCost = i
		} else {
			slog.Error("invalid config value", "key", "auth.bcrypt.cost", "value", v, "error", err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.session.ttl.minutes"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthSessionTTLMinutes = i
		} else {
			slog.Error("invalid config value", "key", "auth.session.ttl.minutes", "value", v, "error", err)
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
			slog.Error("invalid config value", "key", "auth.login.backoff.window.minutes", "value", v, "error", err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.login.backoff.threshold"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffThreshold = i
		} else {
			slog.Error("invalid config value", "key", "auth.login.backoff.threshold", "value", v, "error", err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.login.backoff.base.seconds"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffBaseSeconds = i
		} else {
			slog.Error("invalid config value", "key", "auth.login.backoff.base.seconds", "value", v, "error", err)
			return nil, err
		}
	}
	if v, ok := appProps["auth.login.backoff.max.seconds"]; ok {
		if i, err := strconv.Atoi(v); err == nil {
			cfg.AuthLoginBackoffMaxSeconds = i
		} else {
			slog.Error("invalid config value", "key", "auth.login.backoff.max.seconds", "value", v, "error", err)
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
			slog.Error("invalid config value", "key", "oauth.token.refresh.interval.minutes", "value", v, "error", err)
			return nil, err
		}
	}

	// Weather
	if v, ok := appProps["weather.latitude"]; ok {
		cfg.WeatherLatitude = v
	}
	if v, ok := appProps["weather.longitude"]; ok {
		cfg.WeatherLongitude = v
	}
	if v, ok := appProps["weather.location.name"]; ok {
		cfg.WeatherLocationName = v
	}

	// Telemetry
	cfg.LogLevel = appProps["log.level"]

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
		slog.Error("failed to reload configuration", "error", err)
		return
	}

	AppConfig = cfg

	// Don't include the sensitive ones!
	slog.Info("configuration reloaded",
		"sensor_collection_interval", AppConfig.SensorCollectionInterval,
		"sensor_discovery_skip", AppConfig.SensorDiscoverySkip,
		"openapi_location", AppConfig.OpenAPILocation,
		"health_history_retention_days", AppConfig.HealthHistoryRetentionDays,
		"sensor_data_retention_days", AppConfig.SensorDataRetentionDays,
		"data_cleanup_interval_hours", AppConfig.DataCleanupIntervalHours,
		"health_history_default_response_number", AppConfig.HealthHistoryDefaultResponseNumber,
		"failed_login_retention_days", AppConfig.FailedLoginRetentionDays,
		"auth_bcrypt_cost", AppConfig.AuthBcryptCost,
		"auth_session_ttl_minutes", AppConfig.AuthSessionTTLMinutes,
		"auth_session_cookie_name", AppConfig.AuthSessionCookieName,
		"auth_login_backoff_window_minutes", AppConfig.AuthLoginBackoffWindowMinutes,
		"auth_login_backoff_threshold", AppConfig.AuthLoginBackoffThreshold,
		"auth_login_backoff_base_seconds", AppConfig.AuthLoginBackoffBaseSeconds,
		"auth_login_backoff_max_seconds", AppConfig.AuthLoginBackoffMaxSeconds,
		"oauth_credentials_file_path", AppConfig.OAuthCredentialsFilePath,
		"oauth_token_file_path", AppConfig.OAuthTokenFilePath,
		"oauth_token_refresh_interval_minutes", AppConfig.OAuthTokenRefreshIntervalMinutes,
		"smtp_user", AppConfig.SMTPUser,
		"database_path", AppConfig.DatabasePath,
		"weather_latitude", AppConfig.WeatherLatitude,
		"weather_longitude", AppConfig.WeatherLongitude,
		"weather_location_name", AppConfig.WeatherLocationName,
		"log_level", AppConfig.LogLevel,
	)
}
