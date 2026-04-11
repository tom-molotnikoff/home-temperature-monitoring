package api

import (
	"net/http"
	"strconv"

	"example/sensorHub/service"
	"example/sensorHub/types"

	"github.com/gin-gonic/gin"
)

var mqttService service.MQTTServiceInterface

func InitMQTTAPI(s service.MQTTServiceInterface) {
	mqttService = s
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
		brokers = []types.MQTTBroker{}
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

	var broker types.MQTTBroker
	if err := c.BindJSON(&broker); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := mqttService.AddBroker(ctx, broker); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Broker created"})
}

func updateBrokerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid broker ID"})
		return
	}

	var broker types.MQTTBroker
	if err := c.BindJSON(&broker); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	broker.Id = id

	if err := mqttService.UpdateBroker(ctx, broker); err != nil {
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
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Broker deleted"})
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
			subs = []types.MQTTSubscription{}
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
		subs = []types.MQTTSubscription{}
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

	var sub types.MQTTSubscription
	if err := c.BindJSON(&sub); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}

	if err := mqttService.AddSubscription(ctx, sub); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Subscription created"})
}

func updateSubscriptionHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid subscription ID"})
		return
	}

	var sub types.MQTTSubscription
	if err := c.BindJSON(&sub); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	sub.Id = id

	if err := mqttService.UpdateSubscription(ctx, sub); err != nil {
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
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, gin.H{"message": "Subscription deleted"})
}
