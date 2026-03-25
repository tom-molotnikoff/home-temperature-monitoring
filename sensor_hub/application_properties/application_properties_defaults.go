package appProps

var ApplicationPropertiesDefaults = map[string]string{
	"sensor.collection.interval":             "300",
	"openapi.yaml.location":                  "./docker_tests/openapi.yaml",
	"sensor.discovery.skip":                  "true",
	"health.history.retention.days":          "180",
	"sensor.data.retention.days":             "365",
	"failed.login.retention.days":            "2",
	"data.cleanup.interval.hours":            "24",
	"health.history.default.response.number": "5000",
	// Auth defaults
	"auth.bcrypt.cost":                  "12",
	"auth.session.ttl.minutes":          "43200", // 30 days
	"auth.session.cookie.name":          "sensor_hub_session",
	"auth.login.backoff.window.minutes": "15",
	"auth.login.backoff.threshold":      "5",
	"auth.login.backoff.base.seconds":   "2",
	"auth.login.backoff.max.seconds":    "300",
	// OAuth defaults
	"oauth.credentials.file.path":          "credentials.json",
	"oauth.token.file.path":                "token.json",
	"oauth.token.refresh.interval.minutes": "30",
}

var SmtpPropertiesDefaults = map[string]string{
	"smtp.user": "",
}

var DatabasePropertiesDefaults = map[string]string{
	"database.path": "data/sensor_hub.db",
}
