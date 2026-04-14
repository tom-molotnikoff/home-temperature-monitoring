package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseISO8601Duration_ValidDurations(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
	}{
		{"PT10S", 10 * time.Second},
		{"PT1M", 1 * time.Minute},
		{"PT5M", 5 * time.Minute},
		{"PT15M", 15 * time.Minute},
		{"PT1H", 1 * time.Hour},
		{"PT1H30M", 1*time.Hour + 30*time.Minute},
		{"P1D", 24 * time.Hour},
		{"P7D", 7 * 24 * time.Hour},
		{"P30D", 30 * 24 * time.Hour},
		{"P1DT6H", 30 * time.Hour},
		{"pt5m", 5 * time.Minute}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d, err := ParseISO8601Duration(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, d)
		})
	}
}

func TestParseISO8601Duration_InvalidDurations(t *testing.T) {
	tests := []string{
		"",
		"5M",
		"PT",
		"P",
		"invalid",
		"P0D",
		"PT0S",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseISO8601Duration(input)
			assert.Error(t, err)
		})
	}
}

func TestParseAggregationTiers_FromString(t *testing.T) {
	tiers, err := ParseAggregationTiers("PT15M:raw,PT1H:PT10S,PT6H:PT1M,P1D:PT5M")
	require.NoError(t, err)
	require.Len(t, tiers, 4)

	// Should be sorted by MaxSpan ascending
	assert.Equal(t, 15*time.Minute, tiers[0].MaxSpan)
	assert.Equal(t, "raw", tiers[0].Interval)

	assert.Equal(t, 1*time.Hour, tiers[1].MaxSpan)
	assert.Equal(t, "PT10S", tiers[1].Interval)

	assert.Equal(t, 6*time.Hour, tiers[2].MaxSpan)
	assert.Equal(t, "PT1M", tiers[2].Interval)

	assert.Equal(t, 24*time.Hour, tiers[3].MaxSpan)
	assert.Equal(t, "PT5M", tiers[3].Interval)
}

func TestParseAggregationTiers_EmptyReturnsNil(t *testing.T) {
	tiers, err := ParseAggregationTiers("")
	require.NoError(t, err)
	assert.Nil(t, tiers)
}

func TestParseAggregationTiers_InvalidThreshold(t *testing.T) {
	_, err := ParseAggregationTiers("INVALID:PT5M")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tier threshold")
}

func TestParseAggregationTiers_InvalidInterval(t *testing.T) {
	_, err := ParseAggregationTiers("PT1H:bad")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tier interval")
}

func TestParseAggregationTiers_InvalidFormat(t *testing.T) {
	_, err := ParseAggregationTiers("PT1H")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected THRESHOLD:INTERVAL")
}

func TestResolveAggregationInterval_DefaultTiers(t *testing.T) {
	tests := []struct {
		name     string
		span     time.Duration
		expected string
	}{
		{"10min span → raw", 10 * time.Minute, "raw"},
		{"15min exact → raw", 15 * time.Minute, "raw"},
		{"30min span → PT10S", 30 * time.Minute, "PT10S"},
		{"1h exact → PT10S", 1 * time.Hour, "PT10S"},
		{"3h span → PT1M", 3 * time.Hour, "PT1M"},
		{"6h exact → PT1M", 6 * time.Hour, "PT1M"},
		{"12h span → PT5M", 12 * time.Hour, "PT5M"},
		{"24h exact → PT5M", 24 * time.Hour, "PT5M"},
		{"3d span → PT15M", 3 * 24 * time.Hour, "PT15M"},
		{"7d exact → PT15M", 7 * 24 * time.Hour, "PT15M"},
		{"14d span → PT1H", 14 * 24 * time.Hour, "PT1H"},
		{"30d exact → PT1H", 30 * 24 * time.Hour, "PT1H"},
		{"60d span → P1D (fallback)", 60 * 24 * time.Hour, "P1D"},
		{"365d span → P1D (fallback)", 365 * 24 * time.Hour, "P1D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveAggregationInterval(tt.span, DefaultAggregationTiers)
			assert.Equal(t, tt.expected, result)
		})
	}
}
