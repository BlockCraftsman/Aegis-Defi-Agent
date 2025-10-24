package portfolio

import (
	"math/big"
	"testing"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTest creates a test portfolio with logging and monitoring
type testSetup struct {
	logger  logging.Logger
	monitor *monitoring.Monitor
}

func setupTest(t *testing.T) *testSetup {
	logger, err := logging.NewLogger(&config.LoggingConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	})
	require.NoError(t, err)

	monitor := monitoring.NewMonitor(&config.MonitoringConfig{Enabled: false}, logger)

	return &testSetup{
		logger:  logger,
		monitor: monitor,
	}
}

func TestNewPortfolio(t *testing.T) {
	setup := setupTest(t)

	riskProfile := RiskProfile{
		Type:            RiskModerate,
		MaxDrawdown:     20.0,
		MaxPositionSize: 10.0,
		MaxLeverage:     3.0,
		TargetAllocations: map[string]float64{
			"BTC":  40.0,
			"ETH":  30.0,
			"USDT": 30.0,
		},
	}

	portfolio := NewPortfolio("test-portfolio", "Test Portfolio", riskProfile, setup.logger, setup.monitor)

	assert.Equal(t, "test-portfolio", portfolio.ID)
	assert.Equal(t, "Test Portfolio", portfolio.Name)
	assert.Equal(t, riskProfile, portfolio.RiskProfile)
	assert.Equal(t, big.NewFloat(0), portfolio.TotalValue)
	assert.Equal(t, big.NewFloat(0), portfolio.CashBalance)
	assert.NotNil(t, portfolio.Assets)
	assert.NotNil(t, portfolio.Positions)
}

func TestPortfolio_AddAsset(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	asset := &Asset{
		Symbol:     "BTC",
		Name:       "Bitcoin",
		Amount:     big.NewFloat(1.5),
		PriceUSD:   big.NewFloat(50000),
		ValueUSD:   big.NewFloat(75000),
		LastUpdate: time.Now(),
	}

	err := portfolio.AddAsset(asset)
	require.NoError(t, err)

	// Test adding duplicate asset
	err = portfolio.AddAsset(asset)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Verify asset was added
	assert.Len(t, portfolio.Assets, 1)
	assert.Equal(t, asset, portfolio.Assets["BTC"])
}

func TestPortfolio_UpdateAsset(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	asset := &Asset{
		Symbol:     "ETH",
		Name:       "Ethereum",
		Amount:     big.NewFloat(10),
		PriceUSD:   big.NewFloat(3000),
		ValueUSD:   big.NewFloat(30000),
		LastUpdate: time.Now(),
	}

	err := portfolio.AddAsset(asset)
	require.NoError(t, err)

	// Update asset
	newAmount := big.NewFloat(15)
	newPrice := big.NewFloat(3200)
	err = portfolio.UpdateAsset("ETH", newAmount, newPrice)
	require.NoError(t, err)

	updatedAsset := portfolio.Assets["ETH"]
	assert.Equal(t, newAmount, updatedAsset.Amount)
	assert.Equal(t, newPrice, updatedAsset.PriceUSD)
	assert.Equal(t, big.NewFloat(48000), updatedAsset.ValueUSD) // 15 * 3200

	// Test updating non-existent asset
	err = portfolio.UpdateAsset("BTC", newAmount, newPrice)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortfolio_RemoveAsset(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	asset := &Asset{
		Symbol: "SOL",
		Name:   "Solana",
		Amount: big.NewFloat(100),
	}

	err := portfolio.AddAsset(asset)
	require.NoError(t, err)
	assert.Len(t, portfolio.Assets, 1)

	// Remove asset
	err = portfolio.RemoveAsset("SOL")
	require.NoError(t, err)
	assert.Len(t, portfolio.Assets, 0)

	// Test removing non-existent asset
	err = portfolio.RemoveAsset("BTC")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortfolio_OpenClosePosition(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	position := &Position{
		ID:           "pos-1",
		Asset:        "BTC",
		Type:         PositionLong,
		Size:         big.NewFloat(1.0),
		EntryPrice:   big.NewFloat(50000),
		CurrentPrice: big.NewFloat(50000),
		Leverage:     1.0,
	}

	// Open position
	err := portfolio.OpenPosition(position)
	require.NoError(t, err)
	assert.Len(t, portfolio.Positions, 1)
	assert.Equal(t, PositionOpen, portfolio.Positions["pos-1"].Status)

	// Test opening duplicate position
	err = portfolio.OpenPosition(position)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Close position
	closedPosition, err := portfolio.ClosePosition("pos-1", big.NewFloat(55000))
	require.NoError(t, err)
	assert.Equal(t, PositionClosed, closedPosition.Status)
	assert.Equal(t, big.NewFloat(5000), closedPosition.Pnl) // (55000 - 50000) * 1.0
	assert.Equal(t, 10.0, closedPosition.PnlPercent)        // (5000 / 50000) * 100

	// Test closing non-existent position
	_, err = portfolio.ClosePosition("pos-2", big.NewFloat(55000))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestPortfolio_GetTotalValue(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	// Add cash
	portfolio.CashBalance = big.NewFloat(10000)

	// Add assets
	assets := []*Asset{
		{
			Symbol:     "BTC",
			Amount:     big.NewFloat(0.5),
			PriceUSD:   big.NewFloat(50000),
			ValueUSD:   big.NewFloat(25000),
			LastUpdate: time.Now(),
		},
		{
			Symbol:     "ETH",
			Amount:     big.NewFloat(10),
			PriceUSD:   big.NewFloat(3000),
			ValueUSD:   big.NewFloat(30000),
			LastUpdate: time.Now(),
		},
	}

	for _, asset := range assets {
		err := portfolio.AddAsset(asset)
		require.NoError(t, err)
	}

	totalValue := portfolio.GetTotalValue()
	expectedTotal := big.NewFloat(65000) // 10000 + 25000 + 30000
	assert.Equal(t, expectedTotal, totalValue)
}

func TestPortfolio_GetAssetAllocation(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	// Add cash
	portfolio.CashBalance = big.NewFloat(20000)

	// Add assets
	assets := []*Asset{
		{
			Symbol:     "BTC",
			Amount:     big.NewFloat(1.0),
			PriceUSD:   big.NewFloat(50000),
			ValueUSD:   big.NewFloat(50000),
			LastUpdate: time.Now(),
		},
		{
			Symbol:     "ETH",
			Amount:     big.NewFloat(10),
			PriceUSD:   big.NewFloat(3000),
			ValueUSD:   big.NewFloat(30000),
			LastUpdate: time.Now(),
		},
	}

	for _, asset := range assets {
		err := portfolio.AddAsset(asset)
		require.NoError(t, err)
	}

	allocation := portfolio.GetAssetAllocation()

	// Total value = 20000 + 50000 + 30000 = 100000
	expectedAllocation := map[string]float64{
		"BTC":  50.0, // 50000 / 100000 * 100
		"ETH":  30.0, // 30000 / 100000 * 100
		"CASH": 20.0, // 20000 / 100000 * 100
	}

	assert.InDeltaMapValues(t, expectedAllocation, allocation, 0.1)
}

func TestPortfolio_GetOpenPositions(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	positions := []*Position{
		{
			ID:     "pos-1",
			Asset:  "BTC",
			Type:   PositionLong,
			Size:   big.NewFloat(1.0),
			Status: PositionOpen,
		},
		{
			ID:     "pos-2",
			Asset:  "ETH",
			Type:   PositionShort,
			Size:   big.NewFloat(10.0),
			Status: PositionOpen,
		},
		{
			ID:     "pos-3",
			Asset:  "SOL",
			Type:   PositionLong,
			Size:   big.NewFloat(100.0),
			Status: PositionClosed,
		},
	}

	for _, position := range positions {
		portfolio.Positions[position.ID] = position
	}

	openPositions := portfolio.GetOpenPositions()
	assert.Len(t, openPositions, 2)

	// Verify only open positions are returned
	for _, position := range openPositions {
		assert.Equal(t, PositionOpen, position.Status)
	}
}

func TestPortfolio_GetPortfolioStats(t *testing.T) {
	setup := setupTest(t)

	portfolio := NewPortfolio("test", "Test", RiskProfile{}, setup.logger, setup.monitor)

	// Add cash and assets
	portfolio.CashBalance = big.NewFloat(10000)

	asset := &Asset{
		Symbol:     "BTC",
		Amount:     big.NewFloat(1.0),
		PriceUSD:   big.NewFloat(50000),
		ValueUSD:   big.NewFloat(50000),
		LastUpdate: time.Now(),
	}
	err := portfolio.AddAsset(asset)
	require.NoError(t, err)

	// Add positions
	position := &Position{
		ID:     "pos-1",
		Asset:  "BTC",
		Type:   PositionLong,
		Size:   big.NewFloat(0.5),
		Status: PositionOpen,
	}
	err = portfolio.OpenPosition(position)
	require.NoError(t, err)

	stats := portfolio.GetPortfolioStats()

	assert.Equal(t, big.NewFloat(60000), stats.TotalValue) // 10000 + 50000
	assert.Equal(t, big.NewFloat(10000), stats.CashBalance)
	assert.Equal(t, 1, stats.AssetCount)
	assert.Equal(t, 1, stats.OpenPositions)
	assert.Equal(t, 1, stats.TotalPositions)
	assert.NotNil(t, stats.AssetAllocation)
	assert.WithinDuration(t, time.Now(), stats.LastUpdate, time.Second)
}
