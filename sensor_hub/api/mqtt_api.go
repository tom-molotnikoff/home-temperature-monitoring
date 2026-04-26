package api

import (
	"net/http"
	"strconv"
	"strings"

	gen "example/sensorHub/gen"
	mqttpkg "example/sensorHub/mqtt"
	"example/sensorHub/service"

	"github.com/gin-gonic/gin"
)

var mqttService service.MQTTServiceInterface
var mqttConnManager MQTTStatsProvider

// MQTTStatsProvider is the subset of ConnectionManager the API layer needs.
type MQTTStatsProvider interface {
	Stats() map[int]mqttpkg.BrokerStats
	IsConnected(brokerID int) bool
}

func InitMQTTAPI(s service.MQTTServiceInterface) {
	mqttService = s
}

// InitMQTTStatsProvider sets the connection manager used by the stats endpoint.
func InitMQTTStatsProvider(p MQTTStatsProvider) {
	mqttConnManager = p
}

// isValidationError returns true if the error is a validation error from the
// service layer rather than an infrastructure failure. This is used to return
// 400 Bad Request instead of 500 Internal Server Error.
func isValidationError(err error) bool {
	msg := err.Error()
	validationPrefixes := []string{
		"broker name", "broker host", "broker port", "broker type", "broker id",
		"topic pattern", "driver type", "driver ", "unknown driver",
		"subscription id", "broker not found",
		"multi-level wildcard",
		"broker host:port",
	}
	for _, prefix := range validationPrefixes {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

func isNotFoundError(err error) bool {
	return strings.Contains(err.Error(), "no MQTT broker found") ||
		strings.Contains(err.Error(), "no MQTT subscription found")
}

func isDuplicateError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") ||
		strings.Contains(msg, "already in use")
}

// ============================================================================
// Broker handlers
// ============================================================================

func listBrokersHandler(c *gin.Context) {
	ctx := c.Request.Context()

	brokers, err := mqttService.GetAllBrokers(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error listing brokers"})
		return
	}
	if brokers == nil {
		brokers = []gen.MQTTBroker{}
	}
	c.IndentedJSON(http.StatusOK, brokers)
}

func getBrokerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
		return
	}

	broker, err := mqttService.GetBrokerByID(ctx, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error getting broker"})
		return
	}
	if broker == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Broker not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, broker)
}

func createBrokerHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var broker gen.MQTTBroker
	if err := c.BindJSON(&broker); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	id, err := mqttService.AddBroker(ctx, broker)
	if err != nil {
		if isDuplicateError(err) {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "A broker with that name already exists"})
			return
		}
		if isValidationError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func updateBrokerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
		return
	}

	var broker gen.MQTTBroker
	if err := c.BindJSON(&broker); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	broker.Id = &id

	if err := mqttService.UpdateBroker(ctx, broker); err != nil {
		if isValidationError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Broker updated"})
}

func deleteBrokerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
		return
	}

	if err := mqttService.DeleteBroker(ctx, id); err != nil {
		if isNotFoundError(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Broker not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================================================================
// Subscription handlers
// ============================================================================

func listSubscriptionsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	// If broker_id query param is set, filter by broker
	if brokerParam := c.Query("broker_id"); brokerParam != "" {
		brokerID, err := strconv.Atoi(brokerParam)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker_id parameter"})
			return
		}
		subs, err := mqttService.GetSubscriptionsByBrokerID(ctx, brokerID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error listing subscriptions"})
			return
		}
		if subs == nil {
			subs = []gen.MQTTSubscription{}
		}
		c.IndentedJSON(http.StatusOK, subs)
		return
	}

	subs, err := mqttService.GetAllSubscriptions(ctx)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error listing subscriptions"})
		return
	}
	if subs == nil {
		subs = []gen.MQTTSubscription{}
	}
	c.IndentedJSON(http.StatusOK, subs)
}

func getSubscriptionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
		return
	}

	sub, err := mqttService.GetSubscriptionByID(ctx, id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error getting subscription"})
		return
	}
	if sub == nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Subscription not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, sub)
}

func createSubscriptionHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var sub gen.MQTTSubscription
	if err := c.BindJSON(&sub); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	id, err := mqttService.AddSubscription(ctx, sub)
	if err != nil {
		if isValidationError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"id": id})
}

func updateSubscriptionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
		return
	}

	var sub gen.MQTTSubscription
	if err := c.BindJSON(&sub); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	sub.Id = &id

	if err := mqttService.UpdateSubscription(ctx, sub); err != nil {
		if isValidationError(err) {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Subscription updated"})
}

func deleteSubscriptionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
		return
	}

	if err := mqttService.DeleteSubscription(ctx, id); err != nil {
		if isNotFoundError(err) {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Subscription not found"})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================================================================
// Stats handler
// ============================================================================

func mqttStatsHandler(c *gin.Context) {
	if mqttConnManager == nil {
		c.IndentedJSON(http.StatusServiceUnavailable, gin.H{"message": "MQTT stats not available"})
		return
	}

	statsMap := mqttConnManager.Stats()

	result := make([]mqttpkg.BrokerStats, 0, len(statsMap))
	for _, bs := range statsMap {
		result = append(result, bs)
	}

	c.IndentedJSON(http.StatusOK, result)
}
