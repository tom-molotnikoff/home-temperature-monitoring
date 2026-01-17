package ws

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type connInfo struct {
	conn   *websocket.Conn
	send   chan any
	topics map[string]bool
}

type Hub struct {
	mu           sync.Mutex
	conns        map[*websocket.Conn]*connInfo
	writeTimeout time.Duration
}

var DefaultHub = NewHub()

func NewHub() *Hub {
	return &Hub{
		conns:        make(map[*websocket.Conn]*connInfo),
		writeTimeout: 5 * time.Second,
	}
}

func (h *Hub) Register(conn *websocket.Conn, topics []string) {
	h.mu.Lock()
	if _, ok := h.conns[conn]; ok {
		h.mu.Unlock()
		return
	}
	send := make(chan any, 16)
	topMap := make(map[string]bool)
	for _, t := range topics {
		if t != "" {
			topMap[t] = true
		}
	}
	ci := &connInfo{conn: conn, send: send, topics: topMap}
	h.conns[conn] = ci
	h.mu.Unlock()

	go h.writePump(ci)
	go h.readPump(ci)
	log.Printf("ws: registered connection (total=%d)", h.Count())
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	ci, ok := h.conns[conn]
	if ok {
		delete(h.conns, conn)
	}
	h.mu.Unlock()
	if ok {
		close(ci.send)
		_ = conn.Close()
		log.Printf("ws: unregistered connection (total=%d)", h.Count())
	}
}

func (h *Hub) Subscribe(conn *websocket.Conn, topic string) {
	if topic == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if ci, ok := h.conns[conn]; ok {
		ci.topics[topic] = true
	}
}

func (h *Hub) Unsubscribe(conn *websocket.Conn, topic string) {
	if topic == "" {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if ci, ok := h.conns[conn]; ok {
		delete(ci.topics, topic)
	}
}

func (h *Hub) BroadcastToTopic(topic string, v any) {
	if topic == "" {
		log.Printf("ws: BroadcastToTopic called with empty topic, ignoring")
		return
	}
	h.mu.Lock()
	for _, ci := range h.conns {
		if ci.topics[topic] {
			select {
			case ci.send <- v:
				// queued
			default:
				log.Printf("ws: dropping message for conn on topic %s (buffer full), unregistering", topic)
				delete(h.conns, ci.conn)
				close(ci.send)
				_ = ci.conn.Close()
			}
		}
	}
	h.mu.Unlock()
}

func (h *Hub) Count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.conns)
}

func (h *Hub) writePump(ci *connInfo) {
	for msg := range ci.send {
		_ = ci.conn.SetWriteDeadline(time.Now().Add(h.writeTimeout))
		if err := ci.conn.WriteJSON(msg); err != nil {
			log.Printf("ws: write error: %v", err)
			h.Unregister(ci.conn)
			return
		}
	}
}

func (h *Hub) readPump(ci *connInfo) {
	for {
		if _, _, err := ci.conn.NextReader(); err != nil {
			h.Unregister(ci.conn)
			return
		}
	}
}

func Register(conn *websocket.Conn, topics []string) {
	DefaultHub.Register(conn, topics)
}

func Unregister(conn *websocket.Conn) {
	DefaultHub.Unregister(conn)
}

func BroadcastToTopic(topic string, v any) {
	DefaultHub.BroadcastToTopic(topic, v)
}

// BroadcastToUser broadcasts a message to a specific user's notification topic
func BroadcastToUser(userID int, v any) {
	topic := fmt.Sprintf("notifications:user:%d", userID)
	DefaultHub.BroadcastToTopic(topic, v)
}

// UserNotificationTopic returns the topic name for a specific user's notifications
func UserNotificationTopic(userID int) string {
	return fmt.Sprintf("notifications:user:%d", userID)
}
