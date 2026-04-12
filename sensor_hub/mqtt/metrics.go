package mqtt

import (
	"example/sensorHub/telemetry"

	"go.opentelemetry.io/otel/metric"
)

// mqttInstruments holds all OTel metric instruments for the MQTT subsystem.
type mqttInstruments struct {
	messagesReceived  metric.Int64Counter
	messageErrors     metric.Int64Counter
	connectionsActive metric.Int64UpDownCounter
	devicesDiscovered metric.Int64Counter
	processingTime    metric.Float64Histogram
}

// newMQTTInstruments creates OTel instruments from the global meter provider.
func newMQTTInstruments() *mqttInstruments {
	meter := telemetry.Meter("mqtt")

	messagesReceived, _ := meter.Int64Counter("mqtt.messages.received",
		metric.WithDescription("Total MQTT messages received"),
		metric.WithUnit("{message}"))

	messageErrors, _ := meter.Int64Counter("mqtt.messages.errors",
		metric.WithDescription("Total MQTT message processing errors"),
		metric.WithUnit("{error}"))

	connectionsActive, _ := meter.Int64UpDownCounter("mqtt.connections.active",
		metric.WithDescription("Number of active MQTT broker connections"),
		metric.WithUnit("{connection}"))

	devicesDiscovered, _ := meter.Int64Counter("mqtt.devices.discovered",
		metric.WithDescription("Total MQTT devices auto-discovered"),
		metric.WithUnit("{device}"))

	processingTime, _ := meter.Float64Histogram("mqtt.message.processing.duration",
		metric.WithDescription("Time to process an MQTT message"),
		metric.WithUnit("ms"))

	return &mqttInstruments{
		messagesReceived:  messagesReceived,
		messageErrors:     messageErrors,
		connectionsActive: connectionsActive,
		devicesDiscovered: devicesDiscovered,
		processingTime:    processingTime,
	}
}
