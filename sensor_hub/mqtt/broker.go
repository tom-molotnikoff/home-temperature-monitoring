// Package mqtt provides an embedded MQTT broker powered by mochi-mqtt.
// The broker runs inside the Sensor Hub process and handles local MQTT
// traffic for push-based sensor drivers (Zigbee2MQTT, rtl_433, etc.).
package mqtt

import (
	"fmt"
	"log/slog"
	"sync"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// BrokerConfig holds the configuration for the embedded MQTT broker.
type BrokerConfig struct {
	TCPAddress string // e.g. ":1883"
}

// EmbeddedBroker wraps a mochi-mqtt server instance with lifecycle management.
type EmbeddedBroker struct {
	server  *mqtt.Server
	config  BrokerConfig
	logger  *slog.Logger
	running bool
	mu      sync.Mutex
}

// NewEmbeddedBroker creates a new embedded broker but does not start it.
func NewEmbeddedBroker(config BrokerConfig, logger *slog.Logger) *EmbeddedBroker {
	return &EmbeddedBroker{
		config: config,
		logger: logger.With("component", "embedded_broker"),
	}
}

// Start initialises the mochi-mqtt server, adds a TCP listener, and begins
// serving. The server runs in a background goroutine; call Stop to shut it down.
func (b *EmbeddedBroker) Start() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.running {
		return fmt.Errorf("embedded broker is already running")
	}

	b.server = mqtt.New(&mqtt.Options{
		InlineClient: true,
	})

	tcp := listeners.NewTCP(listeners.Config{
		ID:      "sensor-hub-tcp",
		Address: b.config.TCPAddress,
	})
	if err := b.server.AddListener(tcp); err != nil {
		return fmt.Errorf("failed to add TCP listener on %s: %w", b.config.TCPAddress, err)
	}

	go func() {
		if err := b.server.Serve(); err != nil {
			b.logger.Error("embedded MQTT broker error", "error", err)
		}
	}()

	b.running = true
	b.logger.Info("embedded MQTT broker started", "address", b.config.TCPAddress)
	return nil
}

// Stop gracefully shuts down the embedded broker.
func (b *EmbeddedBroker) Stop() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.running {
		return nil
	}

	if err := b.server.Close(); err != nil {
		return fmt.Errorf("failed to stop embedded broker: %w", err)
	}

	b.running = false
	b.logger.Info("embedded MQTT broker stopped")
	return nil
}

// IsRunning returns whether the broker is currently running.
func (b *EmbeddedBroker) IsRunning() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.running
}

// Server returns the underlying mochi-mqtt server instance. This is useful
// for the connection manager to subscribe to the embedded broker directly
// via the inline client, bypassing the network.
func (b *EmbeddedBroker) Server() *mqtt.Server {
	return b.server
}
