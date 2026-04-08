package types

type Reading struct {
	Id              int      `json:"id"`
	SensorName      string   `json:"sensor_name"`
	MeasurementType string   `json:"measurement_type"`
	NumericValue    *float64 `json:"numeric_value"`
	TextState       *string  `json:"text_state"`
	Unit            string   `json:"unit"`
	Time            string   `json:"time"`
}
