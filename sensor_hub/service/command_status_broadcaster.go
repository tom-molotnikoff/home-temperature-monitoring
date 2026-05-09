package service

import "example/sensorHub/ws"

type websocketCommandStatusBroadcaster struct{}

func (websocketCommandStatusBroadcaster) BroadcastCommandStatus(message CommandStatusMessage) {
	ws.BroadcastToTopic("current-readings", message)
}
