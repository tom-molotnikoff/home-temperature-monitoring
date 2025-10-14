package integration

import (
	"io"
	"net/http"
	"testing"
)

func TestGetAllSensors(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/temperature/sensors/collect")
	if err != nil {
		t.Fatalf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Response: %s", string(body))
}

func TestGetSpecificSensor(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/temperature/sensors/collect/Upstairs")
	if err != nil {
		t.Fatalf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Response: %s", string(body))
}

func TestGetSpecificSensor_UnknownSensor(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/temperature/sensors/collect/UnknownSensor")
	if err != nil {
		t.Fatalf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Response: %s", string(body))
}
