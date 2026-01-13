package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPropertiesFile(t *testing.T) {
	props, err := ReadPropertiesFile("testdata/valid.properties")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if props["key1"] != "value1" || props["key2"] != "value2" {
		t.Errorf("unexpected properties: %v", props)
	}

	_, err = ReadPropertiesFile("testdata/nonexistent.properties")
	if err == nil {
		t.Fatalf("expected error for non-existent file, got nil")
	}

	props, err = ReadPropertiesFile("testdata/invalid.properties")
	if err != nil {
		t.Fatalf("expected no error for invalid format, got %v", err)
	}
	if len(props) != 0 {
		t.Errorf("expected empty properties for invalid format, got %v", props)
	}
}

func TestReadPropertiesFile_ValidFile(t *testing.T) {
	props, err := ReadPropertiesFile("testdata/valid.properties")
	assert.NoError(t, err)
	assert.Equal(t, "value1", props["key1"])
	assert.Equal(t, "value2", props["key2"])
	assert.Equal(t, "value3", props["key3"])
	assert.Equal(t, 3, len(props))
}

func TestReadPropertiesFile_NonExistentFile(t *testing.T) {
	props, err := ReadPropertiesFile("testdata/nonexistent.properties")
	assert.Error(t, err)
	assert.Nil(t, props)
	assert.Contains(t, err.Error(), "failed to open properties file")
}

func TestReadPropertiesFile_InvalidFormat(t *testing.T) {
	props, err := ReadPropertiesFile("testdata/invalid.properties")
	assert.NoError(t, err)
	assert.Equal(t, 0, len(props))
}

func TestNormalizeTimeToSpaceFormat_RFC3339Nano(t *testing.T) {
	input := "2024-01-15T10:30:45.123456789Z"
	expected := "2024-01-15 10:30:45"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_RFC3339(t *testing.T) {
	input := "2024-01-15T10:30:45Z"
	expected := "2024-01-15 10:30:45"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_DateTimeWithoutTimezone(t *testing.T) {
	input := "2024-01-15T10:30:45"
	expected := "2024-01-15 10:30:45"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_DateTimeWithSpace(t *testing.T) {
	input := "2024-01-15 10:30:45"
	expected := "2024-01-15 10:30:45"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_DateOnly(t *testing.T) {
	input := "2024-01-15"
	expected := "2024-01-15 00:00:00"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_UnixTimestamp(t *testing.T) {
	input := "1705315845"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Contains(t, result, "2024-01-15")
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, result)
}

func TestNormalizeTimeToSpaceFormat_EmptyString(t *testing.T) {
	input := ""
	expected := ""
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_InvalidFormat(t *testing.T) {
	input := "invalid-date-format"
	expected := "invalid-date-format"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_RandomString(t *testing.T) {
	input := "hello world"
	expected := "hello world"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}

func TestNormalizeTimeToSpaceFormat_AlmostValidTimestamp(t *testing.T) {
	input := "not-a-timestamp-123abc"
	expected := "not-a-timestamp-123abc"
	result := NormalizeTimeToSpaceFormat(input)
	assert.Equal(t, expected, result)
}
