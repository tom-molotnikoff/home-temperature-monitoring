package appProps

var ApplicationPropertiesDefaults = map[string]string{
	"email.alert.high.temperature.threshold": "28",
	"email.alert.low.temperature.threshold":  "10",
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
}

var SmtpPropertiesDefaults = map[string]string{
	"smtp.user":      "",
	"smtp.recipient": "",
}

var DatabasePropertiesDefaults = map[string]string{
	"database.username": "root",
	"database.password": "password",
	"database.hostname": "mysql",
	"database.port":     "3306",
}
