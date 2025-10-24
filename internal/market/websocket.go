package market

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketService handles real-time market data via WebSocket
type WebSocketService struct {
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	marketData *Data
	isRunning  bool
	stopChan   chan struct{}
	broadcast  chan []byte
}

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Time    time.Time   `json:"time"`
}

// PriceUpdate represents a real-time price update
type PriceUpdate struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
}

// ProtocolUpdate represents a real-time protocol update
type ProtocolUpdate struct {
	Name      string  `json:"name"`
	TVL       float64 `json:"tvl"`
	APY       float64 `json:"apy"`
	Category  string  `json:"category"`
	Timestamp int64   `json:"timestamp"`
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(marketData *Data) *WebSocketService {
	return &WebSocketService{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				// In production, you should validate the origin
				return true
			},
		},
		clients:    make(map[*websocket.Conn]bool),
		marketData: marketData,
		isRunning:  false,
		stopChan:   make(chan struct{}),
		broadcast:  make(chan []byte, 256),
	}
}

// Start begins the WebSocket service
func (ws *WebSocketService) Start(port int) error {
	if ws.isRunning {
		return fmt.Errorf("WebSocket service is already running")
	}

	ws.isRunning = true

	// Start the broadcast handler
	go ws.broadcastHandler()

	// Start the market data update loop
	go ws.marketDataUpdateLoop()

	// Set up HTTP handler
	http.HandleFunc("/ws", ws.handleWebSocket)
	http.HandleFunc("/health", ws.handleHealth)

	log.Printf("WebSocket service starting on port %d", port)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

// Stop halts the WebSocket service
func (ws *WebSocketService) Stop() {
	if !ws.isRunning {
		return
	}

	ws.isRunning = false
	close(ws.stopChan)

	// Close all client connections
	ws.clientsMu.Lock()
	for client := range ws.clients {
		client.Close()
	}
	ws.clients = make(map[*websocket.Conn]bool)
	ws.clientsMu.Unlock()

	log.Println("WebSocket service stopped")
}

// handleWebSocket handles WebSocket connections
func (ws *WebSocketService) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	ws.clientsMu.Lock()
	ws.clients[conn] = true
	ws.clientsMu.Unlock()

	log.Printf("New WebSocket client connected: %s", r.RemoteAddr)

	// Send initial market data
	ws.sendInitialData(conn)

	// Handle client messages
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		ws.handleClientMessage(conn, messageType, message)
	}

	// Unregister client
	ws.clientsMu.Lock()
	delete(ws.clients, conn)
	ws.clientsMu.Unlock()

	log.Printf("WebSocket client disconnected: %s", r.RemoteAddr)
}

// handleHealth handles health check requests
func (ws *WebSocketService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"clients":   len(ws.clients),
		"running":   ws.isRunning,
		"timestamp": time.Now().Unix(),
	})
}

// sendInitialData sends initial market data to a new client
func (ws *WebSocketService) sendInitialData(conn *websocket.Conn) {
	// Send all current prices
	prices := ws.marketData.GetAllPrices()
	priceUpdates := make([]PriceUpdate, 0, len(prices))

	for symbol, priceData := range prices {
		priceUpdates = append(priceUpdates, PriceUpdate{
			Symbol:    symbol,
			Price:     priceData.Price,
			Change24h: priceData.Change24h,
			Volume:    priceData.Volume,
			Timestamp: time.Now().Unix(),
		})
	}

	initialMessage := WebSocketMessage{
		Type: "initial_prices",
		Payload: map[string]interface{}{
			"prices": priceUpdates,
		},
		Time: time.Now(),
	}

	ws.sendMessage(conn, initialMessage)

	// Send all protocols
	protocols := ws.marketData.GetProtocols()
	protocolUpdates := make([]ProtocolUpdate, 0, len(protocols))

	for _, protocol := range protocols {
		protocolUpdates = append(protocolUpdates, ProtocolUpdate{
			Name:      protocol.Name,
			TVL:       protocol.TVL,
			APY:       protocol.APY,
			Category:  protocol.Category,
			Timestamp: time.Now().Unix(),
		})
	}

	protocolMessage := WebSocketMessage{
		Type: "initial_protocols",
		Payload: map[string]interface{}{
			"protocols": protocolUpdates,
		},
		Time: time.Now(),
	}

	ws.sendMessage(conn, protocolMessage)
}

// handleClientMessage processes messages from clients
func (ws *WebSocketService) handleClientMessage(conn *websocket.Conn, messageType int, message []byte) {
	if messageType != websocket.TextMessage {
		return
	}

	var msg WebSocketMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Failed to parse client message: %v", err)
		return
	}

	switch msg.Type {
	case "ping":
		// Respond to ping
		pongMsg := WebSocketMessage{
			Type:    "pong",
			Payload: "pong",
			Time:    time.Now(),
		}
		ws.sendMessage(conn, pongMsg)

	case "subscribe":
		// Handle subscription requests
		ws.handleSubscription(conn, msg.Payload)

	case "unsubscribe":
		// Handle unsubscription requests
		ws.handleUnsubscription(conn, msg.Payload)

	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleSubscription handles client subscription requests
func (ws *WebSocketService) handleSubscription(conn *websocket.Conn, payload interface{}) {
	// For now, we'll just acknowledge the subscription
	// In a more advanced implementation, you could track specific subscriptions
	ackMsg := WebSocketMessage{
		Type:    "subscription_ack",
		Payload: "subscribed",
		Time:    time.Now(),
	}
	ws.sendMessage(conn, ackMsg)
}

// handleUnsubscription handles client unsubscription requests
func (ws *WebSocketService) handleUnsubscription(conn *websocket.Conn, payload interface{}) {
	// For now, we'll just acknowledge the unsubscription
	ackMsg := WebSocketMessage{
		Type:    "unsubscription_ack",
		Payload: "unsubscribed",
		Time:    time.Now(),
	}
	ws.sendMessage(conn, ackMsg)
}

// broadcastHandler sends messages to all connected clients
func (ws *WebSocketService) broadcastHandler() {
	for {
		select {
		case message := <-ws.broadcast:
			ws.clientsMu.RLock()
			for client := range ws.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Broadcast error: %v", err)
					client.Close()
					delete(ws.clients, client)
				}
			}
			ws.clientsMu.RUnlock()

		case <-ws.stopChan:
			return
		}
	}
}

// marketDataUpdateLoop continuously updates and broadcasts market data
func (ws *WebSocketService) marketDataUpdateLoop() {
	ticker := time.NewTicker(5 * time.Second) // Update every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !ws.isRunning {
				return
			}

			// Update market data
			ws.marketData.UpdatePrices()

			// Broadcast price updates
			ws.broadcastPriceUpdates()

			// Broadcast protocol updates (less frequently)
			if time.Now().Second()%30 == 0 { // Every 30 seconds
				ws.broadcastProtocolUpdates()
			}

		case <-ws.stopChan:
			return
		}
	}
}

// broadcastPriceUpdates broadcasts current prices to all clients
func (ws *WebSocketService) broadcastPriceUpdates() {
	prices := ws.marketData.GetAllPrices()
	priceUpdates := make([]PriceUpdate, 0, len(prices))

	for symbol, priceData := range prices {
		priceUpdates = append(priceUpdates, PriceUpdate{
			Symbol:    symbol,
			Price:     priceData.Price,
			Change24h: priceData.Change24h,
			Volume:    priceData.Volume,
			Timestamp: time.Now().Unix(),
		})
	}

	message := WebSocketMessage{
		Type: "price_update",
		Payload: map[string]interface{}{
			"prices": priceUpdates,
		},
		Time: time.Now(),
	}

	ws.broadcastMessage(message)
}

// broadcastProtocolUpdates broadcasts protocol data to all clients
func (ws *WebSocketService) broadcastProtocolUpdates() {
	protocols := ws.marketData.GetProtocols()
	protocolUpdates := make([]ProtocolUpdate, 0, len(protocols))

	for _, protocol := range protocols {
		protocolUpdates = append(protocolUpdates, ProtocolUpdate{
			Name:      protocol.Name,
			TVL:       protocol.TVL,
			APY:       protocol.APY,
			Category:  protocol.Category,
			Timestamp: time.Now().Unix(),
		})
	}

	message := WebSocketMessage{
		Type: "protocol_update",
		Payload: map[string]interface{}{
			"protocols": protocolUpdates,
		},
		Time: time.Now(),
	}

	ws.broadcastMessage(message)
}

// sendMessage sends a message to a specific client
func (ws *WebSocketService) sendMessage(conn *websocket.Conn, message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// broadcastMessage sends a message to all connected clients
func (ws *WebSocketService) broadcastMessage(message WebSocketMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}

	select {
	case ws.broadcast <- data:
		// Message queued for broadcast
	default:
		log.Println("Broadcast channel full, dropping message")
	}
}

// GetClientCount returns the number of connected clients
func (ws *WebSocketService) GetClientCount() int {
	ws.clientsMu.RLock()
	defer ws.clientsMu.RUnlock()
	return len(ws.clients)
}

// IsRunning returns whether the WebSocket service is running
func (ws *WebSocketService) IsRunning() bool {
	return ws.isRunning
}
