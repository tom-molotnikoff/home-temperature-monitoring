package types

// AggregationInterval represents the time-bucket size applied to readings.
// "raw" means no aggregation; otherwise an ISO 8601 duration like "PT5M".
type AggregationInterval string

const (
	AggregationRaw   AggregationInterval = "raw"
	AggregationPT10S AggregationInterval = "PT10S"
	AggregationPT1M  AggregationInterval = "PT1M"
	AggregationPT5M  AggregationInterval = "PT5M"
	AggregationPT15M AggregationInterval = "PT15M"
	AggregationPT1H  AggregationInterval = "PT1H"
	AggregationP1D   AggregationInterval = "P1D"
)

// AggregationFunction represents the SQL aggregate function applied within each bucket.
type AggregationFunction string

const (
	AggregationFunctionNone  AggregationFunction = "none"
	AggregationFunctionAvg   AggregationFunction = "avg"
	AggregationFunctionCount AggregationFunction = "count"
	AggregationFunctionLast  AggregationFunction = "last"
)

// AggregatedReadingsResponse wraps a slice of readings with metadata about the
// aggregation that was applied.
type AggregatedReadingsResponse struct {
	AggregationInterval AggregationInterval `json:"aggregation_interval"`
	AggregationFunction AggregationFunction `json:"aggregation_function"`
	Readings            []Reading           `json:"readings"`
}

// MeasurementTypeAggregation describes which aggregation functions are
// available for a given measurement type.
type MeasurementTypeAggregation struct {
	MeasurementType    string   `json:"measurement_type"`
	DefaultFunction    string   `json:"default_function"`
	SupportedFunctions []string `json:"supported_functions"`
}

const TableMeasurementTypeAggregations = "measurement_type_aggregations"
