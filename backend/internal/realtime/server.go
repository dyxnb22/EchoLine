package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/echoline/echoline/backend/internal/auth"
	"github.com/echoline/echoline/backend/internal/conversation"
	"github.com/echoline/echoline/backend/internal/delivery"
	"github.com/echoline/echoline/backend/internal/message"
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

// conversationGateway supports fanout and ack flows over WS.
type conversationGateway interface {
	ListMemberUserIDs(ctx context.Context, conversationID uuid.UUID) ([]uuid.UUID, error)
	IsMember(ctx context.Context, conversationID, userID uuid.UUID) (bool, error)
	MarkRead(ctx context.Context, conversationID, userID uuid.UUID, seq int64) error
}

// Server handles websocket upgrades and lifecycle.
type Server struct {
	auth          *auth.Service
	hub           *Hub
	messages      *message.Service
	conversations conversationGateway
	deliveries    *delivery.Repository
	presence      PresenceTracker
	logger        *slog.Logger
}

// PresenceTracker optionally records online state.
type PresenceTracker interface {
	Online(ctx context.Context, userID, deviceID string) error
	Offline(ctx context.Context, userID, deviceID string) error
	Refresh(ctx context.Context, userID, deviceID string) error
}

// NewServer creates a realtime websocket server.
func NewServer(
	authSvc *auth.Service,
	messages *message.Service,
	conversations conversationGateway,
	deliveries *delivery.Repository,
	presence PresenceTracker,
	logger *slog.Logger,
) *Server {
	s := &Server{
		auth:          authSvc,
		hub:           NewHub(),
		messages:      messages,
		conversations: conversations,
		deliveries:    deliveries,
		presence:      presence,
		logger:        logger,
	}
	if messages != nil {
		messages.SetBroadcaster(s)
	}
	return s
}

// Hub returns the connection hub.
func (s *Server) Hub() *Hub {
	return s.hub
}

// BroadcastMessageCreated implements message.Broadcaster.
func (s *Server) BroadcastMessageCreated(ctx context.Context, convID uuid.UUID, msg *message.Message, excludeSender bool, senderID uuid.UUID) error {
	memberIDs, err := s.conversations.ListMemberUserIDs(ctx, convID)
	if err != nil {
		return err
	}

	payload := MessageCreatedPayload{
		ID:             msg.ID.String(),
		ConversationID: msg.ConversationID.String(),
		SenderID:       msg.SenderID.String(),
		ClientMsgID:    msg.ClientMsgID,
		Seq:            msg.Seq,
		Type:           string(msg.Type),
		Body:           msg.Body,
		CreatedAt:      msg.CreatedAt.UTC().Format(time.RFC3339),
	}
	raw, err := marshalEnvelope("message.created", "", payload)
	if err != nil {
		return err
	}

	for _, userID := range memberIDs {
		if excludeSender && userID == senderID {
			continue
		}
		s.hub.PushToUser(ctx, userID, raw)
	}
	return nil
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
	if s.presence != nil {
		_ = s.presence.Online(r.Context(), claims.UserID.String(), deviceID)
	}
	s.logger.Info("websocket connected", "user_id", claims.UserID, "device_id", deviceID)

	go conn.writePump()
	conn.readPump(s)
}

func (c *Connection) readPump(s *Server) {
	defer func() {
		s.hub.Unregister(c.UserID, c.DeviceID)
		if s.presence != nil {
			_ = s.presence.Offline(context.Background(), c.UserID.String(), c.DeviceID)
		}
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

		env, err := parseEnvelope(data)
		if err != nil {
			c.sendError("", "invalid_request", err.Error())
			continue
		}

		switch env.Type {
		case "ping":
			var payload PingPayload
			_ = json.Unmarshal(env.Payload, &payload)
			if s.presence != nil {
				_ = s.presence.Refresh(context.Background(), c.UserID.String(), c.DeviceID)
			}
			resp, _ := marshalEnvelope("pong", env.RequestID, PongPayload{TS: payload.TS})
			c.enqueue(resp)
		case "message.send":
			s.handleMessageSend(c, env)
		case "message.ack":
			s.handleMessageAck(c, env)
		default:
			c.sendError(env.RequestID, "unknown_type", "unsupported message type")
		}
	}
}

func (s *Server) handleMessageSend(c *Connection, env Envelope) {
	var payload MessageSendPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		c.sendError(env.RequestID, "invalid_request", "invalid message.send payload")
		return
	}

	convID, err := uuid.Parse(payload.ConversationID)
	if err != nil {
		c.sendError(env.RequestID, "invalid_request", "invalid conversation_id")
		return
	}

	msg, err := s.messages.Send(context.Background(), convID, c.UserID, message.SendInput{
		ClientMsgID: payload.ClientMsgID,
		Type:        message.Type(payload.Type),
		Body:        payload.Body,
		ObjectKey:   payload.AttachmentObjectKey,
	})
	if err != nil {
		if errors.Is(err, conversation.ErrNotMember) {
			c.sendError(env.RequestID, "forbidden", "not a conversation member")
			return
		}
		c.sendError(env.RequestID, "invalid_request", err.Error())
		return
	}

	created := MessageCreatedPayload{
		ID:             msg.ID.String(),
		ConversationID: msg.ConversationID.String(),
		SenderID:       msg.SenderID.String(),
		ClientMsgID:    msg.ClientMsgID,
		Seq:            msg.Seq,
		Type:           string(msg.Type),
		Body:           msg.Body,
		CreatedAt:      msg.CreatedAt.UTC().Format(time.RFC3339),
	}
	resp, _ := marshalEnvelope("message.created", env.RequestID, created)
	c.enqueue(resp)
}

func (s *Server) handleMessageAck(c *Connection, env Envelope) {
	var payload MessageAckPayload
	if err := json.Unmarshal(env.Payload, &payload); err != nil {
		c.sendError(env.RequestID, "invalid_request", "invalid message.ack payload")
		return
	}

	convID, err := uuid.Parse(payload.ConversationID)
	if err != nil {
		c.sendError(env.RequestID, "invalid_request", "invalid conversation_id")
		return
	}
	msgID, err := uuid.Parse(payload.MessageID)
	if err != nil {
		c.sendError(env.RequestID, "invalid_request", "invalid message_id")
		return
	}

	status := delivery.Status(payload.Status)
	if status != delivery.StatusDelivered && status != delivery.StatusRead {
		c.sendError(env.RequestID, "invalid_request", "invalid ack status")
		return
	}

	member, err := s.conversations.IsMember(context.Background(), convID, c.UserID)
	if err != nil || !member {
		c.sendError(env.RequestID, "forbidden", "not a conversation member")
		return
	}

	rec, err := s.deliveries.UpsertACK(context.Background(), msgID, c.UserID, c.DeviceID, status)
	if err != nil {
		if errors.Is(err, delivery.ErrInvalidTransition) {
			c.sendError(env.RequestID, "invalid_transition", "status cannot move backward")
			return
		}
		c.sendError(env.RequestID, "internal_error", "failed to record ack")
		return
	}

	if status == delivery.StatusRead && payload.Seq > 0 {
		_ = s.conversations.MarkRead(context.Background(), convID, c.UserID, payload.Seq)
	}

	resp, _ := marshalEnvelope("message.ack", env.RequestID, map[string]any{
		"message_id": rec.MessageID.String(),
		"status":     rec.Status,
		"acked_at":   rec.AckedAt,
	})
	c.enqueue(resp)
}

func (c *Connection) sendError(requestID, code, message string) {
	raw, err := marshalError(requestID, code, message)
	if err != nil {
		return
	}
	c.enqueue(raw)
}

func (c *Connection) enqueue(raw []byte) {
	select {
	case c.send <- raw:
	default:
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
