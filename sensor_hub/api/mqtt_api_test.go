package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"example/sensorHub/types"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// Mock MQTT service
// ============================================================================

type mockMQTTService struct{ mock.Mock }

func (m *mockMQTTService) AddBroker(ctx context.Context, broker types.MQTTBroker) error {
	return m.Called(ctx, broker).Error(0)
}
func (m *mockMQTTService) GetBrokerByID(ctx context.Context, id int) (*types.MQTTBroker, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MQTTBroker), args.Error(1)
}
func (m *mockMQTTService) GetBrokerByName(ctx context.Context, name string) (*types.MQTTBroker, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MQTTBroker), args.Error(1)
}
func (m *mockMQTTService) GetAllBrokers(ctx context.Context) ([]types.MQTTBroker, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.MQTTBroker), args.Error(1)
}
func (m *mockMQTTService) GetEnabledBrokers(ctx context.Context) ([]types.MQTTBroker, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.MQTTBroker), args.Error(1)
}
func (m *mockMQTTService) UpdateBroker(ctx context.Context, broker types.MQTTBroker) error {
	return m.Called(ctx, broker).Error(0)
}
func (m *mockMQTTService) DeleteBroker(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockMQTTService) AddSubscription(ctx context.Context, sub types.MQTTSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *mockMQTTService) GetSubscriptionByID(ctx context.Context, id int) (*types.MQTTSubscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MQTTSubscription), args.Error(1)
}
func (m *mockMQTTService) GetAllSubscriptions(ctx context.Context) ([]types.MQTTSubscription, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.MQTTSubscription), args.Error(1)
}
func (m *mockMQTTService) GetSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	args := m.Called(ctx, brokerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.MQTTSubscription), args.Error(1)
}
func (m *mockMQTTService) GetEnabledSubscriptionsByBrokerID(ctx context.Context, brokerID int) ([]types.MQTTSubscription, error) {
	args := m.Called(ctx, brokerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.MQTTSubscription), args.Error(1)
}
func (m *mockMQTTService) UpdateSubscription(ctx context.Context, sub types.MQTTSubscription) error {
	return m.Called(ctx, sub).Error(0)
}
func (m *mockMQTTService) DeleteSubscription(ctx context.Context, id int) error {
	return m.Called(ctx, id).Error(0)
}

// ============================================================================
// Test helpers
// ============================================================================

func setupMQTTRouter(method, path string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	apiGroup := router.Group("/api")
	apiGroup.Handle(method, path, handler)
	return router
}

func newMQTTMock() *mockMQTTService {
	m := new(mockMQTTService)
	mqttService = m
	return m
}

// ============================================================================
// Broker tests
// ============================================================================

func TestListBrokersHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	expected := []types.MQTTBroker{{Id: 1, Name: "b1"}, {Id: 2, Name: "b2"}}
	svc.On("GetAllBrokers", mock.Anything).Return(expected, nil)

	router := setupMQTTRouter("GET", "/mqtt/brokers", listBrokersHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/brokers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "b1")
	svc.AssertExpectations(t)
}

func TestListBrokersHandler_Empty(t *testing.T) {
	svc := newMQTTMock()
	svc.On("GetAllBrokers", mock.Anything).Return(nil, nil)

	router := setupMQTTRouter("GET", "/mqtt/brokers", listBrokersHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/brokers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestListBrokersHandler_Error(t *testing.T) {
	svc := newMQTTMock()
	svc.On("GetAllBrokers", mock.Anything).Return(nil, errors.New("db error"))

	router := setupMQTTRouter("GET", "/mqtt/brokers", listBrokersHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/brokers", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetBrokerHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	broker := &types.MQTTBroker{Id: 1, Name: "test-broker", Host: "mqtt.local", Port: 1883}
	svc.On("GetBrokerByID", mock.Anything, 1).Return(broker, nil)

	router := setupMQTTRouter("GET", "/mqtt/brokers/:id", getBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/brokers/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test-broker")
}

func TestGetBrokerHandler_NotFound(t *testing.T) {
	svc := newMQTTMock()
	svc.On("GetBrokerByID", mock.Anything, 99).Return(nil, nil)

	router := setupMQTTRouter("GET", "/mqtt/brokers/:id", getBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/brokers/99", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetBrokerHandler_InvalidID(t *testing.T) {
	_ = newMQTTMock()

	router := setupMQTTRouter("GET", "/mqtt/brokers/:id", getBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/brokers/abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBrokerHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	svc.On("AddBroker", mock.Anything, mock.AnythingOfType("types.MQTTBroker")).Return(nil)

	body, _ := json.Marshal(types.MQTTBroker{Name: "new-broker", Type: "external", Host: "mqtt.local", Port: 1883})
	router := setupMQTTRouter("POST", "/mqtt/brokers", createBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mqtt/brokers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCreateBrokerHandler_InvalidBody(t *testing.T) {
	_ = newMQTTMock()

	router := setupMQTTRouter("POST", "/mqtt/brokers", createBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mqtt/brokers", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBrokerHandler_ServiceError(t *testing.T) {
	svc := newMQTTMock()
	svc.On("AddBroker", mock.Anything, mock.AnythingOfType("types.MQTTBroker")).Return(errors.New("validation failed"))

	body, _ := json.Marshal(types.MQTTBroker{Name: "bad"})
	router := setupMQTTRouter("POST", "/mqtt/brokers", createBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mqtt/brokers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "validation failed")
}

func TestUpdateBrokerHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	svc.On("UpdateBroker", mock.Anything, mock.AnythingOfType("types.MQTTBroker")).Return(nil)

	body, _ := json.Marshal(types.MQTTBroker{Name: "updated", Type: "external", Host: "mqtt.local", Port: 1883})
	router := setupMQTTRouter("PUT", "/mqtt/brokers/:id", updateBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/mqtt/brokers/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestDeleteBrokerHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	svc.On("DeleteBroker", mock.Anything, 1).Return(nil)

	router := setupMQTTRouter("DELETE", "/mqtt/brokers/:id", deleteBrokerHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/mqtt/brokers/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

// ============================================================================
// Subscription tests
// ============================================================================

func TestListSubscriptionsHandler_All(t *testing.T) {
	svc := newMQTTMock()
	expected := []types.MQTTSubscription{{Id: 1, TopicPattern: "zigbee2mqtt/+"}}
	svc.On("GetAllSubscriptions", mock.Anything).Return(expected, nil)

	router := setupMQTTRouter("GET", "/mqtt/subscriptions", listSubscriptionsHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/subscriptions", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "zigbee2mqtt")
}

func TestListSubscriptionsHandler_ByBroker(t *testing.T) {
	svc := newMQTTMock()
	expected := []types.MQTTSubscription{{Id: 1, BrokerId: 2, TopicPattern: "rtl_433/+"}}
	svc.On("GetSubscriptionsByBrokerID", mock.Anything, 2).Return(expected, nil)

	router := setupMQTTRouter("GET", "/mqtt/subscriptions", listSubscriptionsHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/subscriptions?broker_id=2", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "rtl_433")
}

func TestListSubscriptionsHandler_InvalidBrokerID(t *testing.T) {
	_ = newMQTTMock()

	router := setupMQTTRouter("GET", "/mqtt/subscriptions", listSubscriptionsHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/subscriptions?broker_id=abc", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetSubscriptionHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	sub := &types.MQTTSubscription{Id: 1, TopicPattern: "zigbee2mqtt/+", DriverType: "mqtt-zigbee2mqtt"}
	svc.On("GetSubscriptionByID", mock.Anything, 1).Return(sub, nil)

	router := setupMQTTRouter("GET", "/mqtt/subscriptions/:id", getSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/subscriptions/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "zigbee2mqtt")
}

func TestGetSubscriptionHandler_NotFound(t *testing.T) {
	svc := newMQTTMock()
	svc.On("GetSubscriptionByID", mock.Anything, 99).Return(nil, nil)

	router := setupMQTTRouter("GET", "/mqtt/subscriptions/:id", getSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/mqtt/subscriptions/99", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCreateSubscriptionHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	svc.On("AddSubscription", mock.Anything, mock.AnythingOfType("types.MQTTSubscription")).Return(nil)

	body, _ := json.Marshal(types.MQTTSubscription{BrokerId: 1, TopicPattern: "zigbee2mqtt/+", DriverType: "mqtt-zigbee2mqtt"})
	router := setupMQTTRouter("POST", "/mqtt/subscriptions", createSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mqtt/subscriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	svc.AssertExpectations(t)
}

func TestCreateSubscriptionHandler_ServiceError(t *testing.T) {
	svc := newMQTTMock()
	svc.On("AddSubscription", mock.Anything, mock.AnythingOfType("types.MQTTSubscription")).Return(errors.New("driver not found"))

	body, _ := json.Marshal(types.MQTTSubscription{BrokerId: 1, TopicPattern: "test/+", DriverType: "bad"})
	router := setupMQTTRouter("POST", "/mqtt/subscriptions", createSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mqtt/subscriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "driver not found")
}

func TestUpdateSubscriptionHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	svc.On("UpdateSubscription", mock.Anything, mock.AnythingOfType("types.MQTTSubscription")).Return(nil)

	body, _ := json.Marshal(types.MQTTSubscription{BrokerId: 1, TopicPattern: "updated/+", DriverType: "mqtt-zigbee2mqtt"})
	router := setupMQTTRouter("PUT", "/mqtt/subscriptions/:id", updateSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/api/mqtt/subscriptions/1", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestDeleteSubscriptionHandler_Success(t *testing.T) {
	svc := newMQTTMock()
	svc.On("DeleteSubscription", mock.Anything, 1).Return(nil)

	router := setupMQTTRouter("DELETE", "/mqtt/subscriptions/:id", deleteSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", "/api/mqtt/subscriptions/1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestValidateTopicPattern_ViaAPI(t *testing.T) {
	svc := newMQTTMock()
	svc.On("AddSubscription", mock.Anything, mock.AnythingOfType("types.MQTTSubscription")).Return(errors.New("topic pattern must not contain spaces"))

	body, _ := json.Marshal(types.MQTTSubscription{BrokerId: 1, TopicPattern: "bad topic", DriverType: "mqtt-test"})
	router := setupMQTTRouter("POST", "/mqtt/subscriptions", createSubscriptionHandler)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/mqtt/subscriptions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "topic pattern must not contain spaces")
}
