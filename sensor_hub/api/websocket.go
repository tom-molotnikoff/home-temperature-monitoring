package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func createWebSocket(ctx *gin.Context, methodToCall func() (any, error), intervalSeconds int) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}
	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}(conn)

	log.Printf("WebSocket connection established")

	intervalDuration := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(intervalDuration)
	defer ticker.Stop()

	done := make(chan struct{})

	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error (likely closed by client): %v", err)
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
				log.Printf("WebSocket closed or error: %v", err)
				return
			}
		case <-done:
			log.Printf("WebSocket connection closed by client")
			return
		}
	}
}
