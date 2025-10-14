package utils

import (
	"testing"
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
