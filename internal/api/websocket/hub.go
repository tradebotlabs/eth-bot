package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/eth-trading/internal/orchestrator"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a WebSocket client
type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Debug().Str("clientID", client.ID).Msg("WebSocket client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Debug().Str("clientID", client.ID).Msg("WebSocket client disconnected")

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client buffer full, close connection
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all clients
func (h *Hub) Broadcast(msg orchestrator.BroadcastMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal broadcast message")
		return
	}

	select {
	case h.broadcast <- data:
	default:
		log.Warn().Msg("Broadcast channel full, message dropped")
	}
}

// GetClientCount returns number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Close closes all client connections
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for client := range h.clients {
		close(client.Send)
		client.Conn.Close()
		delete(h.clients, client)
	}
}

// HandleConnection handles a new WebSocket connection
func HandleConnection(c echo.Context, hub *Hub, orch *orchestrator.Orchestrator) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to upgrade WebSocket connection")
		return err
	}

	client := &Client{
		ID:   c.Request().RemoteAddr,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  hub,
	}

	hub.register <- client

	// Send initial state in the same format as broadcast
	if orch != nil {
		state := orch.GetState()
		msg := orchestrator.BroadcastMessage{
			Type: orchestrator.MessageTypeState,
			Data: orchestrator.StateUpdate{
				State:   state,
				Summary: nil, // Summary will be populated in subsequent broadcasts
			},
		}
		data, _ := json.Marshal(msg)
		client.Send <- data
	}

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()

	return nil
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().Err(err).Msg("WebSocket read error")
			}
			break
		}

		// Handle incoming messages (e.g., subscription requests)
		c.handleMessage(message)
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Error().Err(err).Msg("WebSocket write error")
			return
		}
	}

	// Hub closed the channel
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// handleMessage handles incoming WebSocket messages
func (c *Client) handleMessage(message []byte) {
	var msg struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data"`
	}

	if err := json.Unmarshal(message, &msg); err != nil {
		log.Error().Err(err).Msg("Failed to parse WebSocket message")
		return
	}

	switch msg.Type {
	case "subscribe":
		// Handle subscription request
		log.Debug().Str("clientID", c.ID).Msg("Client subscribed")
	case "ping":
		// Respond with pong - use select to avoid panic on closed channel
		pong, _ := json.Marshal(map[string]string{"type": "pong"})
		select {
		case c.Send <- pong:
		default:
			// Channel closed or full, ignore
		}
	default:
		log.Debug().Str("type", msg.Type).Msg("Unknown message type")
	}
}
