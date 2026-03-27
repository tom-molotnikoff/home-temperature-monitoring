//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoint(t *testing.T) {
	resp, status := client.GetJSON("/api/health")
	assert.Equal(t, http.StatusOK, status)

	var body map[string]string
	require.NoError(t, json.Unmarshal(resp, &body))
	assert.Equal(t, "ok", body["status"])
}
