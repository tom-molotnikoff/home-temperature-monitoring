package ws

import (
	"log/slog"

	"example/sensorHub/actuation"
)

type CommandStatusBroadcaster struct {
	logger *slog.Logger
}

func NewCommandStatusBroadcaster(logger *slog.Logger) *CommandStatusBroadcaster {
	if logger == nil {
		logger = slog.Default()
	}
	return &CommandStatusBroadcaster{logger: logger.With("component", "command_status_broadcaster")}
}

func (b *CommandStatusBroadcaster) BroadcastCommandStatus(message actuation.CommandStatusMessage) {
	BroadcastToTopic("current-readings", message)
}
