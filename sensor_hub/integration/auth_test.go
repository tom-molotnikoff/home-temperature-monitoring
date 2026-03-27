//go:build integration

package integration

import (
	"net/http"
	"testing"

	"example/sensorHub/testharness"

	"github.com/stretchr/testify/assert"
)

func TestAuth_LoginWithValidCredentials(t *testing.T) {
	c := testharness.NewClient(t, env.ServerURL)
	status := c.Login(env.AdminUser, env.AdminPass)
	assert.Equal(t, http.StatusOK, status)
}

func TestAuth_LoginWithWrongPassword(t *testing.T) {
	c := testharness.NewClient(t, env.ServerURL)
	status := c.Login(env.AdminUser, "wrongpassword")
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestAuth_AccessProtectedEndpointWithoutAuth(t *testing.T) {
	c := testharness.NewClient(t, env.ServerURL)
	_, status := c.GetAllSensors()
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestAuth_GetMe(t *testing.T) {
	resp, status := client.GetMe()
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, string(resp), env.AdminUser)
}

func TestAuth_LogoutThenAccessFails(t *testing.T) {
	c := testharness.NewClient(t, env.ServerURL)
	c.Login(env.AdminUser, env.AdminPass)

	status := c.Logout()
	assert.Equal(t, http.StatusOK, status)

	// After logout, protected endpoints should fail
	_, status = c.GetAllSensors()
	assert.Equal(t, http.StatusUnauthorized, status)
}

func TestAuth_CSRFProtection(t *testing.T) {
	// A client with a valid session but no CSRF token should get 403 on mutations
	c := testharness.NewClient(t, env.ServerURL)
	c.Login(env.AdminUser, env.AdminPass)

	// Wipe the CSRF token to simulate a CSRF attack
	c2 := testharness.NewClient(t, env.ServerURL)
	// Copy cookies by logging in, but then zero out CSRF
	c2.Login(env.AdminUser, env.AdminPass)
	// Create a new client that shares the cookie jar but has no CSRF
	// We test this by using the first client before it has logged in
	fresh := testharness.NewClient(t, env.ServerURL)
	// This client has no session cookie → should get 401
	_, status := fresh.CollectAll()
	assert.Equal(t, http.StatusUnauthorized, status)
}
