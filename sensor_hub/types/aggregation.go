package types

import (
	"time"

	gen "example/sensorHub/gen"
)

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
	Readings            []gen.Reading       `json:"readings"`
}

// MeasurementTypeAggregation describes which aggregation functions are
// available for a given measurement type.
type MeasurementTypeAggregation struct {
	MeasurementType    string   `json:"measurement_type"`
	DefaultFunction    string   `json:"default_function"`
	SupportedFunctions []string `json:"supported_functions"`
}

const TableMeasurementTypeAggregations = "measurement_type_aggregations"

// ErrUnsupportedAggregationFunction is returned when a requested aggregation
// function is not supported for the given measurement type.
type ErrUnsupportedAggregationFunction struct {
	Function        string
	MeasurementType string
	Supported       []string
}

func (e *ErrUnsupportedAggregationFunction) Error() string {
	return "aggregation function \"" + e.Function + "\" is not supported for measurement type \"" + e.MeasurementType + "\"; supported: " + joinStrings(e.Supported)
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}

// ============================================================================
// Aggregation tiers
// ============================================================================

// AggregationTier maps a maximum time-range span to the aggregation bucket size.
type AggregationTier struct {
	MaxSpan  time.Duration
	Interval string // "raw" or ISO 8601 duration like "PT5M"
}
