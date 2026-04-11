// Package mqtt provides the MQTT connection manager that maintains per-broker
// Paho MQTT client connections, manages subscriptions, and routes incoming
// messages to the appropriate PushDriver for parsing.
//
// The ConnectionManager is the bridge between external MQTT brokers and the
// Sensor Hub's driver/service layer. It handles:
//   - Per-broker client lifecycle (connect, reconnect, disconnect)
//   - Subscription management (subscribe/unsubscribe based on DB config)
//   - Message routing: topic → subscription → driver → readings
//   - Auto-discovery: unknown devices become pending sensors

package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	database "example/sensorHub/db"
	"example/sensorHub/drivers"
	"example/sensorHub/service"
	"example/sensorHub/types"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// MessageHandler is called when an MQTT message is received. It is responsible
// for routing the message through the correct driver and storing results.
type MessageHandler func(ctx context.Context, brokerID int, topic string, payload []byte)

// BrokerConnection holds the Paho client and metadata for a single broker.
type BrokerConnection struct {
	Broker types.MQTTBroker
	Client pahomqtt.Client
}

// ConnectionManager manages MQTT client connections for all configured brokers.
type ConnectionManager struct {
	sensorService service.SensorServiceInterface
	readingsRepo  database.ReadingsRepository
	subRepo       database.MQTTSubscriptionRepositoryInterface
	brokerRepo    database.MQTTBrokerRepositoryInterface
	logger        *slog.Logger

	connections map[int]*BrokerConnection // keyed by broker ID
	mu          sync.RWMutex
}

// NewConnectionManager creates a new connection manager.
func NewConnectionManager(
	sensorService service.SensorServiceInterface,
	readingsRepo database.ReadingsRepository,
	subRepo database.MQTTSubscriptionRepositoryInterface,
	brokerRepo database.MQTTBrokerRepositoryInterface,
	logger *slog.Logger,
) *ConnectionManager {
	return &ConnectionManager{
		sensorService: sensorService,
		readingsRepo:  readingsRepo,
		subRepo:       subRepo,
		brokerRepo:    brokerRepo,
		logger:        logger.With("component", "mqtt_connection_manager"),
		connections:   make(map[int]*BrokerConnection),
	}
}

// Start loads all enabled brokers and subscriptions from the database,
// connects to each broker, and subscribes to the configured topics.
func (cm *ConnectionManager) Start(ctx context.Context) error {
	brokers, err := cm.brokerRepo.GetEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to load MQTT brokers: %w", err)
	}

	for _, broker := range brokers {
		if !broker.Enabled {
			cm.logger.Debug("skipping disabled broker", "broker", broker.Name)
			continue
		}
		if broker.Type == "embedded" {
			// Embedded broker connections use localhost
			broker.Host = "localhost"
		}
		if err := cm.ConnectBroker(ctx, broker); err != nil {
			cm.logger.Error("failed to connect to broker", "broker", broker.Name, "error", err)
			continue
		}
	}

	return nil
}

// Stop disconnects all broker clients gracefully.
func (cm *ConnectionManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for id, conn := range cm.connections {
		cm.logger.Info("disconnecting from broker", "broker", conn.Broker.Name)
		conn.Client.Disconnect(250)
		delete(cm.connections, id)
	}
}

// ConnectBroker establishes a connection to the given broker and subscribes
// to all enabled subscriptions for that broker.
func (cm *ConnectionManager) ConnectBroker(ctx context.Context, broker types.MQTTBroker) error {
	brokerURL := fmt.Sprintf("tcp://%s:%d", broker.Host, broker.Port)

	clientID := broker.ClientId
	if clientID == "" {
		clientID = fmt.Sprintf("sensor-hub-%d", broker.Id)
	}

	opts := pahomqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetMaxReconnectInterval(2 * time.Minute).
		SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
			cm.logger.Warn("MQTT connection lost", "broker", broker.Name, "error", err)
		}).
		SetOnConnectHandler(func(client pahomqtt.Client) {
			cm.logger.Info("MQTT connected", "broker", broker.Name)
			// Re-subscribe on reconnect
			go func() {
				if err := cm.subscribeAll(context.Background(), broker.Id, client); err != nil {
					cm.logger.Error("failed to re-subscribe after reconnect", "broker", broker.Name, "error", err)
				}
			}()
		})

	if broker.Username != "" {
		opts.SetUsername(broker.Username)
	}
	if broker.Password != "" {
		opts.SetPassword(broker.Password)
	}

	client := pahomqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("connection to broker %s timed out", broker.Name)
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to connect to broker %s: %w", broker.Name, token.Error())
	}

	cm.mu.Lock()
	cm.connections[broker.Id] = &BrokerConnection{
		Broker: broker,
		Client: client,
	}
	cm.mu.Unlock()

	if err := cm.subscribeAll(ctx, broker.Id, client); err != nil {
		cm.logger.Error("failed to subscribe to topics", "broker", broker.Name, "error", err)
	}

	cm.logger.Info("connected to MQTT broker", "broker", broker.Name, "url", brokerURL)
	return nil
}

// DisconnectBroker disconnects from a specific broker by ID.
func (cm *ConnectionManager) DisconnectBroker(brokerID int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conn, ok := cm.connections[brokerID]
	if !ok {
		return
	}

	conn.Client.Disconnect(250)
	delete(cm.connections, brokerID)
	cm.logger.Info("disconnected from broker", "broker_id", brokerID)
}

// subscribeAll loads enabled subscriptions for a broker and subscribes to each topic.
func (cm *ConnectionManager) subscribeAll(ctx context.Context, brokerID int, client pahomqtt.Client) error {
	subs, err := cm.subRepo.GetEnabledByBrokerID(ctx, brokerID)
	if err != nil {
		return fmt.Errorf("failed to load subscriptions for broker %d: %w", brokerID, err)
	}

	for _, sub := range subs {
		if !sub.Enabled {
			continue
		}
		cm.subscribeTopic(client, brokerID, sub)
	}

	return nil
}

// subscribeTopic subscribes to a single MQTT topic and routes messages.
func (cm *ConnectionManager) subscribeTopic(client pahomqtt.Client, brokerID int, sub types.MQTTSubscription) {
	handler := func(client pahomqtt.Client, msg pahomqtt.Message) {
		cm.handleMessage(context.Background(), brokerID, sub.DriverType, msg.Topic(), msg.Payload())
	}

	token := client.Subscribe(sub.TopicPattern, 0, handler)
	if token.WaitTimeout(5 * time.Second) && token.Error() != nil {
		cm.logger.Error("failed to subscribe", "topic", sub.TopicPattern, "error", token.Error())
		return
	}

	cm.logger.Info("subscribed to MQTT topic", "topic", sub.TopicPattern, "driver", sub.DriverType)
}

// handleMessage processes an incoming MQTT message by routing it through the
// appropriate PushDriver. It handles both known sensors and auto-discovery.
func (cm *ConnectionManager) handleMessage(ctx context.Context, brokerID int, driverType string, topic string, payload []byte) {
	drv, ok := drivers.Get(driverType)
	if !ok {
		cm.logger.Warn("no driver registered for type", "driver", driverType, "topic", topic)
		return
	}

	pushDriver, ok := drv.(drivers.PushDriver)
	if !ok {
		cm.logger.Warn("driver is not a PushDriver", "driver", driverType)
		return
	}

	// Identify the device from the message
	deviceName, err := pushDriver.IdentifyDevice(topic, payload)
	if err != nil {
		cm.logger.Debug("could not identify device", "topic", topic, "error", err)
		return
	}

	// Look up the sensor in the database
	sensor, err := cm.sensorService.ServiceGetSensorByName(ctx, deviceName)
	if err != nil {
		// Sensor not found → auto-discovery: create as pending
		cm.autoDiscoverSensor(ctx, deviceName, driverType, pushDriver)
		return
	}

	// Only process readings for active, enabled sensors
	if sensor.Status != types.SensorStatusActive || !sensor.Enabled {
		return
	}

	readings, err := pushDriver.ParseMessage(topic, payload)
	if err != nil {
		cm.logger.Error("failed to parse MQTT message", "sensor", deviceName, "topic", topic, "error", err)
		cm.sensorService.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorBadHealth,
			fmt.Sprintf("parse error: %v", err))
		return
	}

	if len(readings) == 0 {
		return
	}

	// Tag readings with the sensor name
	for i := range readings {
		readings[i].SensorName = sensor.Name
	}

	if err := cm.readingsRepo.Add(ctx, readings); err != nil {
		cm.logger.Error("failed to store MQTT readings", "sensor", deviceName, "error", err)
		cm.sensorService.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorBadHealth,
			fmt.Sprintf("storage error: %v", err))
		return
	}

	cm.sensorService.ServiceUpdateSensorHealthById(ctx, sensor.Id, types.SensorGoodHealth, "MQTT reading received")
	cm.logger.Debug("processed MQTT message", "sensor", deviceName, "readings", len(readings))
}

// autoDiscoverSensor creates a new sensor in pending status for user approval.
func (cm *ConnectionManager) autoDiscoverSensor(ctx context.Context, deviceName, driverType string, pushDriver drivers.PushDriver) {
	// Check if already exists (race condition guard)
	exists, _ := cm.sensorService.ServiceSensorExists(ctx, deviceName)
	if exists {
		return
	}

	sensor := types.Sensor{
		Name:         deviceName,
		SensorDriver: driverType,
		Config:       map[string]string{},
		Enabled:      false,
		Status:       types.SensorStatusPending,
	}

	if err := cm.sensorService.ServiceAddSensor(ctx, sensor); err != nil {
		cm.logger.Error("failed to auto-discover sensor", "name", deviceName, "error", err)
		return
	}

	cm.logger.Info("auto-discovered MQTT sensor", "name", deviceName, "driver", driverType)
}

// IsConnected returns whether a broker connection is currently active.
func (cm *ConnectionManager) IsConnected(brokerID int) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conn, ok := cm.connections[brokerID]
	if !ok {
		return false
	}
	return conn.Client.IsConnected()
}

// ConnectedBrokerIDs returns the IDs of all currently connected brokers.
func (cm *ConnectionManager) ConnectedBrokerIDs() []int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ids := make([]int, 0, len(cm.connections))
	for id := range cm.connections {
		ids = append(ids, id)
	}
	return ids
}
