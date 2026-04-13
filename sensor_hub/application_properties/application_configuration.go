package appProps

import (
	"log/slog"
	"path/filepath"
)

type ApplicationConfiguration struct {
	SensorCollectionInterval           int    `prop:"sensor.collection.interval" default:"300" file:"application" validate:"positive"`
	SensorDiscoverySkip                bool   `prop:"sensor.discovery.skip" default:"true" file:"application"`
	OpenAPILocation                    string `prop:"openapi.yaml.location" default:"./docker_tests/openapi.yaml" file:"application"`
	HealthHistoryRetentionDays         int    `prop:"health.history.retention.days" default:"30" file:"application" validate:"non_negative"`
	SensorDataRetentionDays            int    `prop:"sensor.data.retention.days" default:"90" file:"application" validate:"non_negative"`
	FailedLoginRetentionDays           int    `prop:"failed.login.retention.days" default:"2" file:"application" validate:"non_negative"`
	DataCleanupIntervalHours           int    `prop:"data.cleanup.interval.hours" default:"1" file:"application" validate:"positive"`
	HealthHistoryDefaultResponseNumber int    `prop:"health.history.default.response.number" default:"1000" file:"application" validate:"positive"`

	SMTPUser string `prop:"smtp.user" default:"" file:"smtp"`

	DatabasePath string `prop:"database.path" default:"data/sensor_hub.db" file:"database" validate:"non_empty"`

	AuthBcryptCost                int    `prop:"auth.bcrypt.cost" default:"12" file:"application"`
	AuthSessionTTLMinutes         int    `prop:"auth.session.ttl.minutes" default:"43200" file:"application"`
	AuthSessionCookieName         string `prop:"auth.session.cookie.name" default:"sensor_hub_session" file:"application"`
	AuthLoginBackoffWindowMinutes int    `prop:"auth.login.backoff.window.minutes" default:"15" file:"application"`
	AuthLoginBackoffThreshold     int    `prop:"auth.login.backoff.threshold" default:"5" file:"application"`
	AuthLoginBackoffBaseSeconds   int    `prop:"auth.login.backoff.base.seconds" default:"2" file:"application"`
	AuthLoginBackoffMaxSeconds    int    `prop:"auth.login.backoff.max.seconds" default:"300" file:"application"`

	OAuthCredentialsFilePath         string `prop:"oauth.credentials.file.path" default:"credentials.json" file:"application"`
	OAuthTokenFilePath               string `prop:"oauth.token.file.path" default:"token.json" file:"application"`
	OAuthTokenRefreshIntervalMinutes int    `prop:"oauth.token.refresh.interval.minutes" default:"30" file:"application"`

	WeatherLatitude     string `prop:"weather.latitude" default:"53.383" file:"application"`
	WeatherLongitude    string `prop:"weather.longitude" default:"-1.4659" file:"application"`
	WeatherLocationName string `prop:"weather.location.name" default:"Sheffield" file:"application"`

	LogLevel string `prop:"log.level" default:"info" file:"application"`

	MQTTBrokerEnabled bool `prop:"mqtt.broker.enabled" default:"true" file:"application"`
	MQTTBrokerPort    int  `prop:"mqtt.broker.port" default:"1883" file:"application" validate:"positive"`
}

var AppConfig *ApplicationConfiguration

func ConvertConfigurationToMaps(cfg *ApplicationConfiguration) (map[string]string, map[string]string, map[string]string) {
	return ConvertToMaps(cfg)
}

func LoadConfigurationFromMaps(appProps, smtpProps, dbProps map[string]string) (*ApplicationConfiguration, error) {
	cfg, err := LoadFromMaps(appProps, smtpProps, dbProps)
	if err != nil {
		return nil, err
	}
	postProcessConfig(cfg)
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

	LogConfig(cfg)
}

// postProcessConfig applies transformations that don't fit into struct tags,
// such as resolving relative OAuth file paths against the config directory.
func postProcessConfig(cfg *ApplicationConfiguration) {
	if cfg.OAuthCredentialsFilePath != "" && !filepath.IsAbs(cfg.OAuthCredentialsFilePath) {
		cfg.OAuthCredentialsFilePath = filepath.Join(configDir, cfg.OAuthCredentialsFilePath)
	}
	if cfg.OAuthTokenFilePath != "" && !filepath.IsAbs(cfg.OAuthTokenFilePath) {
		cfg.OAuthTokenFilePath = filepath.Join(configDir, cfg.OAuthTokenFilePath)
	}
}
