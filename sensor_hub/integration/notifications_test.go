//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifications_ListEmpty(t *testing.T) {
	resp, status := client.GetNotifications(10, 0)
	require.Equal(t, http.StatusOK, status)
	assert.NotNil(t, resp)
}

func TestNotifications_UnreadCount(t *testing.T) {
	resp, status := client.GetUnreadCount()
	require.Equal(t, http.StatusOK, status)

	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(resp, &body))
	// Count should be a number (possibly 0)
	_, ok := body["count"]
	assert.True(t, ok, "response should contain 'count' field")
}

func TestNotifications_BulkMarkAsRead(t *testing.T) {
	status := client.BulkMarkAsRead()
	assert.Equal(t, http.StatusOK, status)
}

func TestNotifications_BulkDismiss(t *testing.T) {
	status := client.BulkDismiss()
	assert.Equal(t, http.StatusOK, status)
}
