package appProps

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// AggregationTier maps a maximum time-range span to the aggregation bucket size.
type AggregationTier struct {
	MaxSpan  time.Duration
	Interval string // "raw" or ISO 8601 duration like "PT5M"
}

// DefaultAggregationTiers are used when no tiers are configured in properties.
var DefaultAggregationTiers = []AggregationTier{
	{MaxSpan: 15 * time.Minute, Interval: "raw"},
	{MaxSpan: 1 * time.Hour, Interval: "PT10S"},
	{MaxSpan: 6 * time.Hour, Interval: "PT1M"},
	{MaxSpan: 24 * time.Hour, Interval: "PT5M"},
	{MaxSpan: 7 * 24 * time.Hour, Interval: "PT15M"},
	{MaxSpan: 30 * 24 * time.Hour, Interval: "PT1H"},
}

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

// ParseAggregationTiers extracts tier config from a properties map.
// It looks for keys prefixed with "readings.aggregation.tier." where the suffix
// is an ISO 8601 duration (the threshold) and the value is either "raw" or an
// ISO 8601 duration (the bucket interval).
func ParseAggregationTiers(props map[string]string) ([]AggregationTier, error) {
	const prefix = "readings.aggregation.tier."
	var tiers []AggregationTier

	for key, val := range props {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		thresholdStr := strings.TrimPrefix(key, prefix)
		threshold, err := ParseISO8601Duration(thresholdStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tier threshold %q: %w", thresholdStr, err)
		}
		interval := strings.TrimSpace(val)
		if interval != "raw" {
			if _, err := ParseISO8601Duration(interval); err != nil {
				return nil, fmt.Errorf("invalid tier interval %q for threshold %q: %w", interval, thresholdStr, err)
			}
		}
		tiers = append(tiers, AggregationTier{MaxSpan: threshold, Interval: interval})
	}

	if len(tiers) == 0 {
		return DefaultAggregationTiers, nil
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
