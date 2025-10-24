package defi

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDeFiAgent(t *testing.T) {
	wallet := &Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := Strategy{
		Type: StrategyArbitrage,
		Parameters: map[string]any{
			"min_profit_threshold": 0.01,
			"max_slippage":         0.005,
		},
		Conditions: []Condition{
			{
				Metric:    "price_difference",
				Operator:  ">",
				Threshold: 0.005,
			},
		},
		IsEnabled: true,
	}

	agent := NewDeFiAgent("test-agent-001", "Test Arbitrage Bot", strategy, wallet)

	assert.Equal(t, "test-agent-001", agent.ID)
	assert.Equal(t, "Test Arbitrage Bot", agent.Name)
	assert.Equal(t, StrategyArbitrage, agent.Strategy.Type)
	assert.False(t, agent.IsActive)
	assert.NotNil(t, agent.MarketData)
	assert.NotNil(t, agent.RiskManager)
	assert.NotNil(t, agent.Blockchain)
}

func TestDeFiAgent_StartStop(t *testing.T) {
	wallet := &Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := Strategy{
		Type:      StrategyArbitrage,
		IsEnabled: true,
	}

	agent := NewDeFiAgent("test-agent-002", "Test Bot", strategy, wallet)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test starting the agent
	err := agent.Start(ctx)
	require.NoError(t, err)
	assert.True(t, agent.IsActive)

	// Give it a moment to start goroutines
	time.Sleep(100 * time.Millisecond)

	// Test stopping the agent
	agent.Stop()
	assert.False(t, agent.IsActive)
}

func TestDeFiAgent_EvaluateConditions(t *testing.T) {
	wallet := &Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := Strategy{
		Type: StrategyArbitrage,
		Conditions: []Condition{
			{
				Metric:    "price_difference",
				Operator:  ">",
				Threshold: 0.03, // Higher than hardcoded 0.02
			},
		},
		IsEnabled: true,
	}

	agent := NewDeFiAgent("test-agent-003", "Test Bot", strategy, wallet)

	// Test condition evaluation
	result := agent.evaluateConditions()
	assert.False(t, result) // Should be false since price difference is hardcoded to 0.02
}

func TestDeFiAgent_GetStatus(t *testing.T) {
	wallet := &Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := Strategy{
		Type:      StrategyYieldFarming,
		IsEnabled: true,
	}

	agent := NewDeFiAgent("test-agent-004", "Yield Farmer", strategy, wallet)
	agent.IsActive = true

	status := agent.GetStatus()

	assert.Equal(t, "test-agent-004", status["id"])
	assert.Equal(t, "Yield Farmer", status["name"])
	assert.Equal(t, StrategyYieldFarming, status["strategy"])
	assert.True(t, status["is_active"].(bool))
	assert.NotNil(t, status["last_activity"])
	assert.NotNil(t, status["risk_score"])
}

func TestCompareValues(t *testing.T) {
	agent := &DeFiAgent{}

	testCases := []struct {
		a        float64
		operator string
		b        float64
		expected bool
	}{
		{1.0, ">", 0.5, true},
		{1.0, "<", 0.5, false},
		{1.0, "==", 1.0, true},
		{1.0, ">=", 1.0, true},
		{1.0, "<=", 1.0, true},
		{1.0, "invalid", 1.0, false},
	}

	for _, tc := range testCases {
		result := agent.compareValues(tc.a, tc.operator, tc.b)
		assert.Equal(t, tc.expected, result, "Failed for %f %s %f", tc.a, tc.operator, tc.b)
	}
}

func TestStrategyTypes(t *testing.T) {
	assert.Equal(t, StrategyType("arbitrage"), StrategyArbitrage)
	assert.Equal(t, StrategyType("yield_farming"), StrategyYieldFarming)
	assert.Equal(t, StrategyType("liquidity_provision"), StrategyLiquidity)
	assert.Equal(t, StrategyType("market_making"), StrategyMarketMaking)
}

func TestRiskManagerDefaults(t *testing.T) {
	riskManager := NewRiskManager()

	assert.Equal(t, 0.1, riskManager.MaxPositionSize)
	assert.Equal(t, 0.005, riskManager.MaxSlippage)
	assert.Equal(t, 0.05, riskManager.StopLossPercent)
	assert.Equal(t, 0.1, riskManager.TakeProfitPercent)
	assert.Equal(t, 0.0, riskManager.RiskScore)
}

func TestMarketDataDefaults(t *testing.T) {
	marketData := NewMarketData()

	assert.NotNil(t, marketData.PythClient)
	assert.NotNil(t, marketData.PriceFeeds)
	assert.Equal(t, 30*time.Second, marketData.UpdateInterval)
}
