package mcpclient

import (
	"math/rand"
	"time"
)

// PythClient provides simulated price data (legacy support)
type PythClient struct{}

type PythPriceData struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Change24h float64 `json:"change_24h"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
}

func NewPythClient() *PythClient {
	return &PythClient{}
}

func (p *PythClient) GetPrice(symbol string) (*PythPriceData, error) {
	// Use realistic mock data
	return p.getRealisticMockPrice(symbol), nil
}

// getRealisticMockPrice generates realistic cryptocurrency prices with small variations
func (p *PythClient) getRealisticMockPrice(symbol string) *PythPriceData {
	basePrices := map[string]PythPriceData{
		"ETH/USD":   {Symbol: "ETH/USD", Price: 3580.25, Change24h: 2.3, Volume: 12.5e9, Timestamp: time.Now().Unix()},
		"BTC/USD":   {Symbol: "BTC/USD", Price: 65250.75, Change24h: 1.7, Volume: 28.3e9, Timestamp: time.Now().Unix()},
		"SOL/USD":   {Symbol: "SOL/USD", Price: 142.80, Change24h: 4.8, Volume: 3.8e9, Timestamp: time.Now().Unix()},
		"USDC/USD":  {Symbol: "USDC/USD", Price: 1.00, Change24h: 0.0, Volume: 6.2e9, Timestamp: time.Now().Unix()},
		"USDT/USD":  {Symbol: "USDT/USD", Price: 0.9997, Change24h: -0.03, Volume: 45.1e9, Timestamp: time.Now().Unix()},
		"MATIC/USD": {Symbol: "MATIC/USD", Price: 0.68, Change24h: -1.2, Volume: 0.9e9, Timestamp: time.Now().Unix()},
	}

	if price, exists := basePrices[symbol]; exists {
		// Add small random variation to simulate real market movement
		variation := (rand.Float64() - 0.5) * 0.02 // ±1% variation
		price.Price = price.Price * (1 + variation)

		// Update timestamp to current time
		price.Timestamp = time.Now().Unix()

		return &price
	}

	// Default fallback
	return &PythPriceData{
		Symbol:    symbol,
		Price:     100 + rand.Float64()*50,
		Change24h: (rand.Float64() - 0.5) * 10,
		Volume:    1e9 + rand.Float64()*5e9,
		Timestamp: time.Now().Unix(),
	}
}

func (p *PythClient) GetPriceWithFallback(symbol string) (*PythPriceData, error) {
	return p.GetPrice(symbol)
}
