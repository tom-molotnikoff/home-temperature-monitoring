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

// DefaultAggregationTiersString is the default value for the readings.aggregation.tiers property.
const DefaultAggregationTiersString = "PT15M:raw,PT1H:PT10S,PT6H:PT1M,P1D:PT5M,P7D:PT15M,P30D:PT1H"

// DefaultAggregationTiers are the parsed default tiers.
var DefaultAggregationTiers = func() []AggregationTier {
	tiers, err := ParseAggregationTiers(DefaultAggregationTiersString)
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
// Example: "PT15M:raw,PT1H:PT10S,PT6H:PT1M"
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
