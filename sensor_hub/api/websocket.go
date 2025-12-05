package api

import (
	"log"
	"net/http"
	"time"

	"example/sensorHub/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func createPushWebSocket(ctx *gin.Context, topic string) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}
	log.Printf("WebSocket connection established (registering to hub) topic=%s", topic)
	ws.Register(conn, []string{topic})
}

func createIntervalBasedWebSocket(ctx *gin.Context, topic string, methodToCall func() (any, error), intervalSeconds int) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	ws.Register(conn, []string{topic})

	intervalDuration := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(intervalDuration)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				close(done)
				return
			}
		}
	}()

	for {
		select {
		case <-ticker.C:
			response, err := methodToCall()
			if err != nil {
				log.Printf("Error fetching latest readings: %v", err)
				continue
			}
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("WebSocket closed or error writing interval message: %v", err)
				ws.Unregister(conn)
				return
			}
		case <-done:
			log.Printf("WebSocket connection closed by client (interval)")
			ws.Unregister(conn)
			return
		}
	}
}
