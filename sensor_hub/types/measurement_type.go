package types

type MeasurementType struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Unit        string `json:"unit"`
	Category    string `json:"category"` // "numeric" or "binary"
}

type SensorMeasurementType struct {
	SensorId          int    `json:"sensor_id"`
	MeasurementTypeId int    `json:"measurement_type_id"`
	MeasurementType   string `json:"measurement_type"`
	Unit              string `json:"unit"`
}
