package market

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"
)

// PythRealClient provides real Pyth Network price data
type PythRealClient struct {
	BaseURL    string
	HTTPClient *http.Client
	PriceFeeds map[string]PythPriceFeed
}

// PythPriceFeed represents a Pyth price feed
type PythPriceFeed struct {
	ID         string  `json:"id"`
	Symbol     string  `json:"symbol"`
	Price      float64 `json:"price"`
	Confidence float64 `json:"confidence"`
	Exponent   int     `json:"exponent"`
	Timestamp  int64   `json:"publish_time"`
	EMA        float64 `json:"ema_price"`
	EMAC       float64 `json:"ema_conf"`
}

// PythResponse represents the Pyth Network API response
type PythResponse struct {
	Parsed []struct {
		ID    string `json:"id"`
		Price struct {
			Price       string `json:"price"`
			Confidence  string `json:"conf"`
			Exponent    int    `json:"exponent"`
			PublishTime int64  `json:"publish_time"`
		} `json:"price"`
	} `json:"parsed"`
}

// NewPythRealClient creates a new Pyth Network client
func NewPythRealClient() *PythRealClient {
	return &PythRealClient{
		BaseURL: "https://hermes.pyth.network/api",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		PriceFeeds: make(map[string]PythPriceFeed),
	}
}

// GetPrice gets real price data from Pyth Network
func (p *PythRealClient) GetPrice(symbol string) (*PythPriceFeed, error) {
	// Map symbols to Pyth price feed IDs
	feedID := p.getFeedID(symbol)
	if feedID == "" {
		return nil, fmt.Errorf("unsupported symbol: %s", symbol)
	}

	url := fmt.Sprintf("%s/latest_price_feeds?ids[]=%s", p.BaseURL, feedID)

	resp, err := p.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch price from Pyth: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pyth API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var pythResp PythResponse
	if err := json.Unmarshal(body, &pythResp); err != nil {
		return nil, fmt.Errorf("failed to parse Pyth response: %v", err)
	}

	if len(pythResp.Parsed) == 0 {
		return nil, fmt.Errorf("no price data found for %s", symbol)
	}

	priceData := pythResp.Parsed[0]

	// Parse price and confidence
	price := new(big.Float)
	price.SetString(priceData.Price.Price)
	priceFloat, _ := price.Float64()

	confidence := new(big.Float)
	confidence.SetString(priceData.Price.Confidence)
	confidenceFloat, _ := confidence.Float64()

	// Apply exponent
	if priceData.Price.Exponent != 0 {
		exponent := float64(priceData.Price.Exponent)
		multiplier := 1.0
		if exponent > 0 {
			for i := 0; i < int(exponent); i++ {
				multiplier *= 10
			}
		} else {
			for i := 0; i < int(-exponent); i++ {
				multiplier /= 10
			}
		}
		priceFloat = priceFloat * multiplier
		confidenceFloat = confidenceFloat * multiplier
	}

	feed := PythPriceFeed{
		ID:         priceData.ID,
		Symbol:     symbol,
		Price:      priceFloat,
		Confidence: confidenceFloat,
		Exponent:   priceData.Price.Exponent,
		Timestamp:  priceData.Price.PublishTime,
	}

	// Cache the price feed
	p.PriceFeeds[symbol] = feed

	log.Printf("Pyth price for %s: %.6f ± %.6f (updated: %s)",
		symbol, priceFloat, confidenceFloat,
		time.Unix(priceData.Price.PublishTime, 0).Format(time.RFC3339))

	return &feed, nil
}

// GetMultiplePrices gets multiple price feeds at once
func (p *PythRealClient) GetMultiplePrices(symbols []string) (map[string]PythPriceFeed, error) {
	results := make(map[string]PythPriceFeed)

	for _, symbol := range symbols {
		feed, err := p.GetPrice(symbol)
		if err != nil {
			log.Printf("Warning: Failed to get price for %s: %v", symbol, err)
			continue
		}
		results[symbol] = *feed
	}

	return results, nil
}

// SubscribeToPriceFeeds starts real-time price updates
func (p *PythRealClient) SubscribeToPriceFeeds(ctx context.Context, symbols []string, updateInterval time.Duration) <-chan map[string]PythPriceFeed {
	priceChan := make(chan map[string]PythPriceFeed)

	go func() {
		ticker := time.NewTicker(updateInterval)
		defer ticker.Stop()
		defer close(priceChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				prices, err := p.GetMultiplePrices(symbols)
				if err != nil {
					log.Printf("Failed to update prices: %v", err)
					continue
				}

				select {
				case priceChan <- prices:
					// Prices sent successfully
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return priceChan
}

// getFeedID maps symbol to Pyth price feed ID
func (p *PythRealClient) getFeedID(symbol string) string {
	// Pyth Network price feed IDs for major cryptocurrencies
	feedIDs := map[string]string{
		"ETH/USD":   "0xff61491a931112ddf1bd8147cd1b641375f79f5825126d665480874634fd0ace",
		"BTC/USD":   "0xe62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43",
		"SOL/USD":   "0xef0d8b6fda2ceba41da15d4095d1da392a0d2f8ed0c6c7bc0f4cfac8c280b56d",
		"USDC/USD":  "0xeaa020c61cc479712813461ce153894a96a6c00b21ed0cfc2798d1f9a9e9c94a",
		"USDT/USD":  "0x2b89b9dc8fdf9f34709a5b106b472f0f39bb6ca9ce04b0fd7f2e971688e2e53b",
		"MATIC/USD": "0x5de33a9112c2b700b8d30b8a3402c103578ccfa2765696471cc672bd5cf6ac52",
		"ARB/USD":   "0x3fa4252848f9f0a1480be62745a4629d9eb1322aebab8a791e344b3b9c1adcf5",
		"LINK/USD":  "0x8ac0c70fff57e9aefdf5edf44b51d62c2d433653cbb2cf5cc06bb115af04d221",
		"AVAX/USD":  "0x93da3352f9f1d105fdfe4971cfa80e9dd777bfc5d0f683ebb6e1294b92137bb7",
		"BNB/USD":   "0x2f95862b045670cd22bee3114c39763a4a08beeb663b145d283c31d7d1101c4f",
	}

	return feedIDs[symbol]
}

// GetPriceWithFallback gets price with fallback to cached data
func (p *PythRealClient) GetPriceWithFallback(symbol string) (*PythPriceFeed, error) {
	// Try to get fresh price
	feed, err := p.GetPrice(symbol)
	if err == nil {
		return feed, nil
	}

	// Fallback to cached data
	if cached, exists := p.PriceFeeds[symbol]; exists {
		log.Printf("Using cached price for %s", symbol)
		return &cached, nil
	}

	return nil, fmt.Errorf("no price data available for %s", symbol)
}

// GetPriceHistory gets historical price data (placeholder for future implementation)
func (p *PythRealClient) GetPriceHistory(symbol string, from, to time.Time) ([]PythPriceFeed, error) {
	// This would require Pyth's historical price API
	// For now, return current price as single data point
	current, err := p.GetPrice(symbol)
	if err != nil {
		return nil, err
	}

	return []PythPriceFeed{*current}, nil
}
