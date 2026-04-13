//go:build integration

package testharness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	"example/sensorHub/types"
)

// Client is an HTTP client for the sensor-hub API with session and CSRF management.
type Client struct {
	t         *testing.T // nil when used from TestMain
	baseURL   string
	http      *http.Client
	csrfToken string
}

// NewClient creates an unauthenticated HTTP client pointed at the test server.
func NewClient(t *testing.T, baseURL string) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		t:       t,
		baseURL: baseURL,
		http:    &http.Client{Jar: jar},
	}
}

func (c *Client) fatalf(format string, args ...any) {
	if c.t != nil {
		c.t.Helper()
		c.t.Fatalf(format, args...)
	} else {
		panic(fmt.Sprintf(format, args...))
	}
}

// --- Auth ---

// Login authenticates and stores the session cookie + CSRF token.
func (c *Client) Login(username, password string) int {
	body := map[string]string{"username": username, "password": password}
	resp, status := c.doRequest("POST", "/api/auth/login", jsonBytes(body))
	if status == http.StatusOK {
		var loginResp struct {
			CSRFToken string `json:"csrf_token"`
		}
		if err := json.Unmarshal(resp, &loginResp); err == nil {
			c.csrfToken = loginResp.CSRFToken
		}
	}
	return status
}

// LoginAdmin logs in with the default admin credentials.
func (c *Client) LoginAdmin(env *Env) {
	status := c.Login(env.AdminUser, env.AdminPass)
	if status != http.StatusOK {
		c.fatalf("admin login failed with status %d", status)
	}
}

func (c *Client) GetMe() (json.RawMessage, int) {
	return c.getJSON("/api/auth/me")
}

func (c *Client) Logout() int {
	_, status := c.doRequest("POST", "/api/auth/logout", nil)
	return status
}

// ChangePassword changes the current user's password.
func (c *Client) ChangePassword(newPassword string) int {
	body := map[string]interface{}{"new_password": newPassword}
	_, status := c.doRequest("PUT", "/api/users/password", jsonBytes(body))
	return status
}

// --- Sensors ---

func (c *Client) AddSensor(sensor types.Sensor) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/sensors/", jsonBytes(sensor))
}

func (c *Client) GetAllSensors() ([]types.Sensor, int) {
	var result []types.Sensor
	status := c.getDecode("/api/sensors/", &result)
	return result, status
}

func (c *Client) GetSensorByName(name string) (types.Sensor, int) {
	var result types.Sensor
	status := c.getDecode("/api/sensors/"+name, &result)
	return result, status
}

func (c *Client) DeleteSensor(name string) int {
	_, status := c.doRequest("DELETE", "/api/sensors/"+name, nil)
	return status
}

func (c *Client) EnableSensor(name string) int {
	_, status := c.doRequest("POST", "/api/sensors/enable/"+name, nil)
	return status
}

func (c *Client) DisableSensor(name string) int {
	_, status := c.doRequest("POST", "/api/sensors/disable/"+name, nil)
	return status
}

func (c *Client) CollectAll() (json.RawMessage, int) {
	return c.doRequest("POST", "/api/sensors/collect", nil)
}

func (c *Client) CollectByName(name string) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/sensors/collect/"+name, nil)
}

// --- Readings ---

func (c *Client) GetReadingsBetween(from, to, sensor string) ([]types.Reading, int) {
	path := fmt.Sprintf("/api/readings/between?start=%s&end=%s", url.QueryEscape(from), url.QueryEscape(to))
	if sensor != "" {
		path += "&sensor=" + url.QueryEscape(sensor)
	}
	var result []types.Reading
	status := c.getDecode(path, &result)
	return result, status
}

func (c *Client) GetHourlyReadings(from, to, sensor string) ([]types.Reading, int) {
	path := fmt.Sprintf("/api/readings/hourly/between?start=%s&end=%s", url.QueryEscape(from), url.QueryEscape(to))
	if sensor != "" {
		path += "&sensor=" + url.QueryEscape(sensor)
	}
	var result []types.Reading
	status := c.getDecode(path, &result)
	return result, status
}

// --- Measurement Types ---

func (c *Client) GetAllMeasurementTypes() (json.RawMessage, int) {
	return c.getJSON("/api/measurement-types")
}

func (c *Client) GetMeasurementTypesWithReadings() (json.RawMessage, int) {
	return c.getJSON("/api/measurement-types?has_readings=true")
}

func (c *Client) GetMeasurementTypesForSensor(sensorID int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/sensors/by-id/%d/measurement-types", sensorID))
}

// --- Alerts ---

type AlertRuleRequest struct {
	SensorID          int     `json:"SensorID"`
	MeasurementTypeId int     `json:"MeasurementTypeId"`
	AlertType         string  `json:"AlertType"`
	HighThreshold     float64 `json:"HighThreshold"`
	LowThreshold      float64 `json:"LowThreshold"`
	RateLimitSeconds    int     `json:"RateLimitSeconds"`
	Enabled           bool    `json:"Enabled"`
}

func (c *Client) CreateAlertRule(rule AlertRuleRequest) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/alerts/", jsonBytes(rule))
}

func (c *Client) GetAlertRulesBySensorID(sensorID int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/alerts/sensor/%d", sensorID))
}

func (c *Client) GetAlertHistory(sensorID int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/alerts/sensor/%d/history", sensorID))
}

// --- Users ---

type CreateUserRequest struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles,omitempty"`
}

func (c *Client) CreateUser(user CreateUserRequest) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/users/", jsonBytes(user))
}

func (c *Client) ListUsers() (json.RawMessage, int) {
	return c.getJSON("/api/users/")
}

func (c *Client) DeleteUser(id int) int {
	_, status := c.doRequest("DELETE", fmt.Sprintf("/api/users/%d", id), nil)
	return status
}

// --- Notifications ---

func (c *Client) GetNotifications(limit, offset int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/notifications/?limit=%d&offset=%d", limit, offset))
}

func (c *Client) GetUnreadCount() (json.RawMessage, int) {
	return c.getJSON("/api/notifications/unread-count")
}

func (c *Client) BulkMarkAsRead() int {
	_, status := c.doRequest("POST", "/api/notifications/bulk/read", nil)
	return status
}

func (c *Client) BulkDismiss() int {
	_, status := c.doRequest("POST", "/api/notifications/bulk/dismiss", nil)
	return status
}

// --- Properties ---

func (c *Client) GetProperties() (json.RawMessage, int) {
	return c.getJSON("/api/properties/")
}

func (c *Client) SetProperty(key, value string) int {
	body := map[string]string{key: value}
	_, status := c.doRequest("PATCH", "/api/properties/", jsonBytes(body))
	return status
}

// --- API Keys ---

func (c *Client) CreateApiKey(name string) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/api-keys/", jsonBytes(map[string]string{"name": name}))
}

// --- Roles ---

func (c *Client) ListRoles() (json.RawMessage, int) {
	return c.getJSON("/api/roles/")
}

// --- Dashboards ---

type CreateDashboardRequest struct {
	Name string `json:"name"`
}

type UpdateDashboardRequest struct {
	Name   string          `json:"name,omitempty"`
	Config json.RawMessage `json:"config,omitempty"`
}

type ShareDashboardRequest struct {
	TargetUserId int `json:"target_user_id"`
}

func (c *Client) ListDashboards() (json.RawMessage, int) {
	return c.getJSON("/api/dashboards/")
}

func (c *Client) GetDashboard(id int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/dashboards/%d", id))
}

func (c *Client) CreateDashboard(name string) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/dashboards/", jsonBytes(CreateDashboardRequest{Name: name}))
}

func (c *Client) UpdateDashboard(id int, req UpdateDashboardRequest) (json.RawMessage, int) {
	return c.doRequest("PUT", fmt.Sprintf("/api/dashboards/%d", id), jsonBytes(req))
}

func (c *Client) DeleteDashboard(id int) int {
	_, status := c.doRequest("DELETE", fmt.Sprintf("/api/dashboards/%d", id), nil)
	return status
}

func (c *Client) ShareDashboard(id, targetUserId int) (json.RawMessage, int) {
	return c.doRequest("POST", fmt.Sprintf("/api/dashboards/%d/share", id), jsonBytes(ShareDashboardRequest{TargetUserId: targetUserId}))
}

func (c *Client) SetDefaultDashboard(id int) (json.RawMessage, int) {
	return c.doRequest("PUT", fmt.Sprintf("/api/dashboards/%d/default", id), jsonBytes(nil))
}

// --- MQTT Brokers ---

func (c *Client) ListMQTTBrokers() (json.RawMessage, int) {
	return c.getJSON("/api/mqtt/brokers")
}

func (c *Client) CreateMQTTBroker(broker types.MQTTBroker) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/mqtt/brokers", jsonBytes(broker))
}

func (c *Client) GetMQTTBroker(id int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/mqtt/brokers/%d", id))
}

func (c *Client) UpdateMQTTBroker(id int, broker types.MQTTBroker) (json.RawMessage, int) {
	return c.doRequest("PUT", fmt.Sprintf("/api/mqtt/brokers/%d", id), jsonBytes(broker))
}

func (c *Client) DeleteMQTTBroker(id int) int {
	_, status := c.doRequest("DELETE", fmt.Sprintf("/api/mqtt/brokers/%d", id), nil)
	return status
}

// --- MQTT Subscriptions ---

func (c *Client) ListMQTTSubscriptions() (json.RawMessage, int) {
	return c.getJSON("/api/mqtt/subscriptions")
}

func (c *Client) CreateMQTTSubscription(sub types.MQTTSubscription) (json.RawMessage, int) {
	return c.doRequest("POST", "/api/mqtt/subscriptions", jsonBytes(sub))
}

func (c *Client) GetMQTTSubscription(id int) (json.RawMessage, int) {
	return c.getJSON(fmt.Sprintf("/api/mqtt/subscriptions/%d", id))
}

func (c *Client) UpdateMQTTSubscription(id int, sub types.MQTTSubscription) (json.RawMessage, int) {
	return c.doRequest("PUT", fmt.Sprintf("/api/mqtt/subscriptions/%d", id), jsonBytes(sub))
}

func (c *Client) DeleteMQTTSubscription(id int) int {
	_, status := c.doRequest("DELETE", fmt.Sprintf("/api/mqtt/subscriptions/%d", id), nil)
	return status
}

// --- Sensor Status ---

func (c *Client) GetSensorsByStatus(status string) (json.RawMessage, int) {
	return c.getJSON("/api/sensors/status/" + url.PathEscape(status))
}

func (c *Client) ApproveSensor(id int) (json.RawMessage, int) {
	return c.doRequest("POST", fmt.Sprintf("/api/sensors/approve/%d", id), nil)
}

func (c *Client) DismissSensor(id int) (json.RawMessage, int) {
	return c.doRequest("POST", fmt.Sprintf("/api/sensors/dismiss/%d", id), nil)
}

// --- Generic helpers (exported for test flexibility) ---

// GetJSON performs a GET request and returns the raw JSON response.
func (c *Client) GetJSON(path string) (json.RawMessage, int) {
	return c.getJSON(path)
}

// --- Internal helpers ---

func (c *Client) getJSON(path string) (json.RawMessage, int) {
	body, status := c.doRequest("GET", path, nil)
	return body, status
}

func (c *Client) getDecode(path string, v interface{}) int {
	body, status := c.doRequest("GET", path, nil)
	if status >= 200 && status < 300 && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			c.fatalf("failed to decode response from %s: %v\nbody: %s", path, err, string(body))
		}
	}
	return status
}

func (c *Client) postJSONDecode(path string, payload interface{}, v interface{}) int {
	body, status := c.doRequest("POST", path, jsonBytes(payload))
	if status >= 200 && status < 300 && len(body) > 0 {
		if err := json.Unmarshal(body, v); err != nil {
			c.fatalf("failed to decode response from %s: %v\nbody: %s", path, err, string(body))
		}
	}
	return status
}

func (c *Client) doRequest(method, path string, body []byte) (json.RawMessage, int) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		c.fatalf("failed to create request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.csrfToken != "" && method != "GET" && method != "HEAD" {
		req.Header.Set("X-CSRF-Token", c.csrfToken)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.fatalf("request %s %s failed: %v", method, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.fatalf("failed to read response body: %v", err)
	}

	return json.RawMessage(respBody), resp.StatusCode
}

func jsonBytes(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}
