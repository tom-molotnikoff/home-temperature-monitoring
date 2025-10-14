package integration

import (
	"io"
	"net/http"
	"testing"
)

func TestGetAllSensors(t *testing.T) {
	resp, err := http.Post("http://localhost:8080/sensors/collect", "", nil)
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
	resp, err := http.Post("http://localhost:8080/sensors/collect/Upstairs", "", nil)
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
	resp, err := http.Post("http://localhost:8080/sensors/collect/UnknownSensor", "", nil)
	if err != nil {
		t.Fatalf("Failed to call API: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("Expected 500 Internal Server Error, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	t.Logf("Response: %s", string(body))
}
