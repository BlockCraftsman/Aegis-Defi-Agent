package market

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/aegis-protocol/aegis-core/pkg/mcpclient"
)

type Data struct {
	Prices     map[string]PriceData
	Protocols  []ProtocolData
	lastUpdate time.Time
	pythClient *mcpclient.PythClient
}

type PriceData struct {
	Symbol    string
	Price     float64
	Change24h float64
	Volume    float64
}

type ProtocolData struct {
	Name     string
	TVL      float64
	APY      float64
	Category string
}

func NewData() *Data {
	data := &Data{
		Prices: map[string]PriceData{
			"ETH":   {Symbol: "ETH", Price: 3500, Change24h: 2.5, Volume: 1.2e9},
			"BTC":   {Symbol: "BTC", Price: 65000, Change24h: 1.8, Volume: 25e9},
			"USDC":  {Symbol: "USDC", Price: 1.00, Change24h: 0.0, Volume: 5.8e9},
			"MATIC": {Symbol: "MATIC", Price: 0.70, Change24h: -0.5, Volume: 0.8e9},
			"SOL":   {Symbol: "SOL", Price: 150, Change24h: 5.2, Volume: 3.2e9},
		},
		Protocols: []ProtocolData{
			{Name: "Uniswap V3", TVL: 4.2e9, APY: 12.5, Category: "DEX"},
			{Name: "Aave V3", TVL: 8.1e9, APY: 3.2, Category: "Lending"},
			{Name: "Compound", TVL: 2.3e9, APY: 2.8, Category: "Lending"},
			{Name: "Curve", TVL: 3.1e9, APY: 4.5, Category: "StableSwap"},
		},
		lastUpdate: time.Now(),
		pythClient: mcpclient.NewPythClient(),
	}

	// Initialize with real Pyth data
	data.updateFromPyth()

	return data
}

func (d *Data) UpdatePrices() {
	// First try to get real prices from Pyth Network
	d.updateFromPyth()

	// If Pyth update failed or we need to simulate some changes
	for symbol, priceData := range d.Prices {
		// Add small random variations to simulate market dynamics
		change := (rand.Float64() - 0.5) * 2 // -1% to +1% small variation
		priceData.Change24h = change
		priceData.Price *= 1 + change/100

		// Simulate volume changes
		volumeChange := (rand.Float64() - 0.3) * 20 // -6% to +14%
		priceData.Volume *= 1 + volumeChange/100

		d.Prices[symbol] = priceData
	}

	// Update protocols
	for i := range d.Protocols {
		// Simulate TVL and APY changes
		tvlChange := (rand.Float64() - 0.2) * 20 // -10% to +10%
		d.Protocols[i].TVL *= 1 + tvlChange/100

		apyChange := (rand.Float64() - 0.5) * 5 // -2.5% to +2.5%
		d.Protocols[i].APY += apyChange
		if d.Protocols[i].APY < 0 {
			d.Protocols[i].APY = 0.1
		}
	}

	d.lastUpdate = time.Now()
}

func (d *Data) GetLastUpdate() time.Time {
	return d.lastUpdate
}

func (d *Data) GetPrice(symbol string) (PriceData, bool) {
	price, exists := d.Prices[symbol]
	return price, exists
}

func (d *Data) GetAllPrices() map[string]PriceData {
	return d.Prices
}

func (d *Data) GetProtocols() []ProtocolData {
	return d.Protocols
}

func (d *Data) updateFromPyth() {
	// Get real prices from Pyth Network
	symbols := []string{"ETH/USD", "BTC/USD", "SOL/USD", "MATIC/USD", "USDC/USD"}

	pythPrices, err := d.pythClient.GetMultiplePrices(symbols)
	if err != nil {
		fmt.Printf("Warning: Failed to fetch Pyth prices: %v\n", err)
		return
	}

	// Update our price data with real Pyth prices
	for pythSymbol, pythData := range pythPrices {
		// Convert Pyth symbol format to our internal format
		var symbol string
		switch pythSymbol {
		case "ETH/USD":
			symbol = "ETH"
		case "BTC/USD":
			symbol = "BTC"
		case "SOL/USD":
			symbol = "SOL"
		case "MATIC/USD":
			symbol = "MATIC"
		case "USDC/USD":
			symbol = "USDC"
		default:
			continue
		}

		if existing, exists := d.Prices[symbol]; exists {
			existing.Price = pythData.Price
			existing.Change24h = pythData.Change24h
			existing.Volume = pythData.Volume
			d.Prices[symbol] = existing
		}
	}
}
