package types

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
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

// DefaultAggregationTiers are the parsed default tiers.
var DefaultAggregationTiers = func() []AggregationTier {
	tiers, err := ParseAggregationTiers("PT15M:raw,PT1H:PT10S,PT6H:PT1M,P1D:PT5M,P7D:PT15M,P30D:PT1H")
	if err != nil {
		panic(fmt.Sprintf("invalid default aggregation tiers: %v", err))
	}
	return tiers
}()

// FallbackInterval is used for queries exceeding the largest tier threshold.
const FallbackInterval = "P1D"

var iso8601Re = regexp.MustCompile(`^P(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$`)

// ParseISO8601Duration parses a subset of ISO 8601 durations: P[n]D, PT[n]H, PT[n]M, PT[n]S
// and combinations (e.g. P1DT6H, PT1H30M).
func ParseISO8601Duration(s string) (time.Duration, error) {
	matches := iso8601Re.FindStringSubmatch(strings.ToUpper(s))
	if matches == nil {
		return 0, fmt.Errorf("invalid ISO 8601 duration: %q", s)
	}
	var d time.Duration
	if matches[1] != "" {
		var days int
		fmt.Sscanf(matches[1], "%d", &days)
		d += time.Duration(days) * 24 * time.Hour
	}
	if matches[2] != "" {
		var hours int
		fmt.Sscanf(matches[2], "%d", &hours)
		d += time.Duration(hours) * time.Hour
	}
	if matches[3] != "" {
		var mins int
		fmt.Sscanf(matches[3], "%d", &mins)
		d += time.Duration(mins) * time.Minute
	}
	if matches[4] != "" {
		var secs int
		fmt.Sscanf(matches[4], "%d", &secs)
		d += time.Duration(secs) * time.Second
	}
	if d == 0 {
		return 0, fmt.Errorf("ISO 8601 duration is zero: %q", s)
	}
	return d, nil
}

// ParseAggregationTiers parses a comma-separated tier string.
// Format: "THRESHOLD:INTERVAL,THRESHOLD:INTERVAL,..." where THRESHOLD and INTERVAL
// are ISO 8601 durations (or "raw" for INTERVAL).
func ParseAggregationTiers(tiersStr string) ([]AggregationTier, error) {
	tiersStr = strings.TrimSpace(tiersStr)
	if tiersStr == "" {
		return nil, nil
	}

	var tiers []AggregationTier
	for _, entry := range strings.Split(tiersStr, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tier entry %q: expected THRESHOLD:INTERVAL", entry)
		}
		thresholdStr := strings.TrimSpace(parts[0])
		interval := strings.TrimSpace(parts[1])

		threshold, err := ParseISO8601Duration(thresholdStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tier threshold %q: %w", thresholdStr, err)
		}
		if interval != "raw" {
			if _, err := ParseISO8601Duration(interval); err != nil {
				return nil, fmt.Errorf("invalid tier interval %q for threshold %q: %w", interval, thresholdStr, err)
			}
		}
		tiers = append(tiers, AggregationTier{MaxSpan: threshold, Interval: interval})
	}

	sort.Slice(tiers, func(i, j int) bool { return tiers[i].MaxSpan < tiers[j].MaxSpan })
	return tiers, nil
}

// ResolveAggregationInterval picks the appropriate interval for a given time span.
func ResolveAggregationInterval(span time.Duration, tiers []AggregationTier) string {
	for _, tier := range tiers {
		if span <= tier.MaxSpan {
			return tier.Interval
		}
	}
	return FallbackInterval
}
