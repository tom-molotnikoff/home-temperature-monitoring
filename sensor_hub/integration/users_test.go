//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"example/sensorHub/testharness"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers_CreateAndList(t *testing.T) {
	user := testharness.CreateUserRequest{
		Username: "integration-viewer",
		Password: "viewerpass123",
		Email:    "viewer@test.com",
	}
	_, status := client.CreateUser(user)
	require.Equal(t, http.StatusCreated, status)

	resp, status := client.ListUsers()
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(resp), "integration-viewer")
}

func TestUsers_ViewerCannotAccessAdminEndpoints(t *testing.T) {
	// Create a viewer user and log in as them
	user := testharness.CreateUserRequest{
		Username: "viewer-restricted",
		Password: "viewerpass456",
		Email:    "restricted@test.com",
	}
	client.CreateUser(user)

	viewer := testharness.NewClient(t, env.ServerURL)
	status := viewer.Login("viewer-restricted", "viewerpass456")
	require.Equal(t, http.StatusOK, status)

	// Viewer should not be able to create sensors (requires manage_sensors)
	_, status = viewer.ListUsers()
	assert.Equal(t, http.StatusForbidden, status)
}

func TestUsers_ListRoles(t *testing.T) {
	resp, status := client.ListRoles()
	require.Equal(t, http.StatusOK, status)

	var roles []json.RawMessage
	require.NoError(t, json.Unmarshal(resp, &roles))
	assert.NotEmpty(t, roles, "should have default roles")
}

func TestUsers_Delete(t *testing.T) {
	user := testharness.CreateUserRequest{
		Username: "user-to-delete",
		Password: "deletepass123",
		Email:    "delete@test.com",
	}
	resp, status := client.CreateUser(user)
	require.Equal(t, http.StatusCreated, status)

	var created struct {
		ID int `json:"id"`
	}
	json.Unmarshal(resp, &created)
	require.True(t, created.ID > 0)

	status = client.DeleteUser(created.ID)
	assert.Equal(t, http.StatusOK, status)
}
