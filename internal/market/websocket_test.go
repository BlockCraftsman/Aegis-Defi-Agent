package market

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewWebSocketService(t *testing.T) {
	marketData := NewData()
	service := NewWebSocketService(marketData)

	if service == nil {
		t.Fatal("Failed to create WebSocket service")
	}

	if service.marketData != marketData {
		t.Error("Market data not properly set")
	}

	if service.clients == nil {
		t.Error("Clients map not initialized")
	}

	if service.broadcast == nil {
		t.Error("Broadcast channel not initialized")
	}
}

func TestWebSocketServiceLifecycle(t *testing.T) {
	marketData := NewData()
	service := NewWebSocketService(marketData)

	// Test starting the service
	go func() {
		// Use a random port for testing
		err := service.Start(0)
		if err != nil && !strings.Contains(err.Error(), "closed network connection") {
			t.Errorf("Failed to start service: %v", err)
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	if !service.IsRunning() {
		t.Error("Service should be running after Start()")
	}

	// Test stopping the service
	service.Stop()

	if service.IsRunning() {
		t.Error("Service should not be running after Stop()")
	}
}

func TestWebSocketConnection(t *testing.T) {
	marketData := NewData()
	service := NewWebSocketService(marketData)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		service.handleWebSocket(w, r)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect to the WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Test that we receive initial data
	// We expect two initial messages: initial_prices and initial_protocols
	for i := 0; i < 2; i++ {
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read initial message %d: %v", i, err)
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			t.Fatalf("Failed to unmarshal message %d: %v", i, err)
		}

		if wsMsg.Type != "initial_prices" && wsMsg.Type != "initial_protocols" {
			t.Errorf("Expected initial_prices or initial_protocols message, got %s", wsMsg.Type)
		}
	}

	// Test ping-pong
	pingMsg := WebSocketMessage{
		Type:    "ping",
		Payload: "ping",
		Time:    time.Now(),
	}

	pingData, _ := json.Marshal(pingMsg)
	if err := conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// Read pong response
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read pong message: %v", err)
	}

	var pongMsg WebSocketMessage
	if err := json.Unmarshal(message, &pongMsg); err != nil {
		t.Fatalf("Failed to unmarshal pong message: %v", err)
	}

	if pongMsg.Type != "pong" {
		t.Errorf("Expected pong message, got %s", pongMsg.Type)
	}
}

func TestHealthEndpoint(t *testing.T) {
	marketData := NewData()
	service := NewWebSocketService(marketData)

	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	service.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal health response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}

	if _, exists := response["clients"]; !exists {
		t.Error("Health response missing clients count")
	}

	if _, exists := response["running"]; !exists {
		t.Error("Health response missing running status")
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("Health response missing timestamp")
	}
}

func TestBroadcastMessage(t *testing.T) {
	marketData := NewData()
	service := NewWebSocketService(marketData)

	// Start the broadcast handler
	go service.broadcastHandler()

	// Create a test message
	testMessage := WebSocketMessage{
		Type:    "test_message",
		Payload: "test_payload",
		Time:    time.Now(),
	}

	// Broadcast the message
	service.broadcastMessage(testMessage)

	// Give the broadcast handler time to process
	time.Sleep(100 * time.Millisecond)

	// The test passes if no panic occurs
}

func TestClientCount(t *testing.T) {
	marketData := NewData()
	service := NewWebSocketService(marketData)

	// Initially should have 0 clients
	if count := service.GetClientCount(); count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}

	// Simulate adding a client
	service.clientsMu.Lock()
	service.clients[&websocket.Conn{}] = true
	service.clientsMu.Unlock()

	// Should now have 1 client
	if count := service.GetClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}
}

func TestPriceUpdateStructure(t *testing.T) {
	priceUpdate := PriceUpdate{
		Symbol:    "ETH",
		Price:     3500.0,
		Change24h: 2.5,
		Volume:    1.2e9,
		Timestamp: time.Now().Unix(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(priceUpdate)
	if err != nil {
		t.Fatalf("Failed to marshal price update: %v", err)
	}

	var decoded PriceUpdate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal price update: %v", err)
	}

	if decoded.Symbol != priceUpdate.Symbol {
		t.Errorf("Symbol mismatch: expected %s, got %s", priceUpdate.Symbol, decoded.Symbol)
	}

	if decoded.Price != priceUpdate.Price {
		t.Errorf("Price mismatch: expected %.2f, got %.2f", priceUpdate.Price, decoded.Price)
	}
}

func TestProtocolUpdateStructure(t *testing.T) {
	protocolUpdate := ProtocolUpdate{
		Name:      "Uniswap V3",
		TVL:       4.2e9,
		APY:       12.5,
		Category:  "DEX",
		Timestamp: time.Now().Unix(),
	}

	// Test JSON marshaling
	data, err := json.Marshal(protocolUpdate)
	if err != nil {
		t.Fatalf("Failed to marshal protocol update: %v", err)
	}

	var decoded ProtocolUpdate
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal protocol update: %v", err)
	}

	if decoded.Name != protocolUpdate.Name {
		t.Errorf("Name mismatch: expected %s, got %s", protocolUpdate.Name, decoded.Name)
	}

	if decoded.TVL != protocolUpdate.TVL {
		t.Errorf("TVL mismatch: expected %.2f, got %.2f", protocolUpdate.TVL, decoded.TVL)
	}
}
