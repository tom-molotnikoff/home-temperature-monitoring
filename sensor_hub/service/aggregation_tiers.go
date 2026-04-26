package service

import (
	"example/sensorHub/utils"
	"fmt"
	"sort"
	"strings"
	"time"
)

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

		threshold, err := utils.ParseISO8601Duration(thresholdStr)
		if err != nil {
			return nil, fmt.Errorf("invalid tier threshold %q: %w", thresholdStr, err)
		}
		if interval != "raw" {
			if _, err := utils.ParseISO8601Duration(interval); err != nil {
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
