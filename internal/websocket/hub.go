package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"telegram-ai-bot/internal/ai"

	"github.com/gorilla/websocket"
)

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mutex      sync.RWMutex
	aiClient   *ai.Client
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID string
}

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
	UserID  string `json:"user_id,omitempty"`
	Stage   string `json:"stage,omitempty"` // thinking, reasoning, responding
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development - in production, you should validate origins
		return true
	},
}

func NewHub(aiClient *ai.Client) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		aiClient:   aiClient,
	}
}

func (h *Hub) Run() {
	for { //nolint:gosimple // This is a message pump that needs to run indefinitely
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("Client registered: %s", client.userID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("Client unregistered: %s", client.userID)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anonymous"
	}

	client := &Client{
		hub:    h,
		conn:   conn,
		send:   make(chan []byte, 256),
		userID: userID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming message and stream AI response
		go c.handleMessage(msg)
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for { //nolint:gosimple // WebSocket message pump needs indefinite loop
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

func (c *Client) handleMessage(msg Message) {
	if msg.Type != "user_message" {
		return
	}

	// Use simple streaming
	c.hub.aiClient.ChatStreamWithThinking(msg.Content, func(stage string, content string, isComplete bool) {
		c.sendMessage(Message{
			Type:    "ai_response",
			Content: content,
			Stage:   stage,
		})
	})
}

func (c *Client) sendMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON marshal error: %v", err)
		return
	}

	select {
	case c.send <- data:
	default:
		close(c.send)
		delete(c.hub.clients, c)
	}
}
