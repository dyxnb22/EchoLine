package realtime

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/echoline/echoline/backend/internal/auth"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 50 * time.Second
	maxMessageSize = 64 * 1024
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Hub tracks active WebSocket connections.
type Hub struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]map[string]*Connection
}

// NewHub creates a connection hub.
func NewHub() *Hub {
	return &Hub{conns: make(map[uuid.UUID]map[string]*Connection)}
}

// Register adds a connection for a user/device pair.
func (h *Hub) Register(userID uuid.UUID, deviceID string, conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.conns[userID]; !ok {
		h.conns[userID] = make(map[string]*Connection)
	}
	h.conns[userID][deviceID] = conn
}

// Unregister removes a connection.
func (h *Hub) Unregister(userID uuid.UUID, deviceID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if devices, ok := h.conns[userID]; ok {
		delete(devices, deviceID)
		if len(devices) == 0 {
			delete(h.conns, userID)
		}
	}
}

// ConnectionCount returns total active connections.
func (h *Hub) ConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, devices := range h.conns {
		total += len(devices)
	}
	return total
}

// Connection wraps a single websocket client.
type Connection struct {
	UserID   uuid.UUID
	DeviceID string
	conn     *websocket.Conn
	send     chan []byte
}

// Server handles websocket upgrades and lifecycle.
type Server struct {
	auth   *auth.Service
	hub    *Hub
	logger *slog.Logger
}

// NewServer creates a realtime websocket server.
func NewServer(authSvc *auth.Service, logger *slog.Logger) *Server {
	return &Server{
		auth:   authSvc,
		hub:    NewHub(),
		logger: logger,
	}
}

// Hub returns the connection hub.
func (s *Server) Hub() *Hub {
	return s.hub
}

// HandleWS upgrades HTTP to websocket after token validation.
func (s *Server) HandleWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	deviceID := r.URL.Query().Get("device_id")
	if token == "" || deviceID == "" {
		http.Error(w, "token and device_id are required", http.StatusUnauthorized)
		return
	}

	claims, err := s.auth.ParseToken(token)
	if err != nil || claims.TokenType != auth.TokenTypeAccess {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("websocket upgrade failed", "error", err)
		return
	}

	conn := &Connection{
		UserID:   claims.UserID,
		DeviceID: deviceID,
		conn:     wsConn,
		send:     make(chan []byte, 16),
	}

	s.hub.Register(claims.UserID, deviceID, conn)
	s.logger.Info("websocket connected", "user_id", claims.UserID, "device_id", deviceID)

	go conn.writePump()
	conn.readPump(s)
}

func (c *Connection) readPump(s *Server) {
	defer func() {
		s.hub.Unregister(c.UserID, c.DeviceID)
		_ = c.conn.Close()
		s.logger.Info("websocket disconnected", "user_id", c.UserID, "device_id", c.DeviceID)
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("websocket read error", "error", err)
			}
			break
		}

		var envelope struct {
			Type    string          `json:"type"`
			Payload json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(data, &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "ping":
			var payload struct {
				TS int64 `json:"ts"`
			}
			_ = json.Unmarshal(envelope.Payload, &payload)
			resp, _ := json.Marshal(map[string]any{
				"type": "pong",
				"payload": map[string]any{
					"ts": payload.TS,
				},
			})
			select {
			case c.send <- resp:
			default:
			}
		}
	}
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// PushToUser sends an event to all devices for a user.
func (h *Hub) PushToUser(ctx context.Context, userID uuid.UUID, payload []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	devices := h.conns[userID]
	for _, conn := range devices {
		select {
		case conn.send <- payload:
		case <-ctx.Done():
			return
		default:
		}
	}
}
