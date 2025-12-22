package appProps

var ApplicationPropertiesDefaults = map[string]string{
	"email.alert.high.temperature.threshold": "28",
	"email.alert.low.temperature.threshold":  "10",
	"sensor.collection.interval":             "300",
	"openapi.yaml.location":                  "./docker_tests/openapi.yaml",
	"sensor.discovery.skip":                  "true",
	"health.history.retention.days":          "180",
	"sensor.data.retention.days":             "365",
	"data.cleanup.interval.hours":            "24",
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
