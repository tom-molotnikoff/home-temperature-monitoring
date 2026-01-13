package types

import "time"

type AlertHistoryEntry struct {
	ID           int       `json:"id"`
	SensorID     int       `json:"sensor_id"`
	AlertType    string    `json:"alert_type"`
	ReadingValue string    `json:"reading_value"`
	SentAt       time.Time `json:"sent_at"`
}
