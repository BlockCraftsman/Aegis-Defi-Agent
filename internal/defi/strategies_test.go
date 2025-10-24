package defi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStrategyEngine(t *testing.T) {
	engine := NewStrategyEngine()

	assert.NotNil(t, engine)
	assert.NotNil(t, engine.Strategies)
	assert.NotNil(t, engine.Agents)
	assert.False(t, engine.IsRunning)
}

func TestStrategyEngine_AddRemoveStrategy(t *testing.T) {
	engine := NewStrategyEngine()

	strategy := &TradingStrategy{
		ID:          "test-strategy-001",
		Name:        "Test Strategy",
		Type:        StrategyArbitrage,
		Description: "Test strategy for unit testing",
		Parameters: StrategyParameters{
			MaxPositionSize: 0.1,
			MinProfitMargin: 0.01,
			MaxSlippage:     0.005,
			ExecutionDelay:  5,
			CooldownPeriod:  60,
			TargetAssets:    []string{"ETH", "USDC"},
		},
		IsActive: true,
	}

	// Test adding strategy
	err := engine.AddStrategy(strategy)
	require.NoError(t, err)
	assert.Len(t, engine.Strategies, 1)
	assert.Equal(t, strategy, engine.Strategies["test-strategy-001"])

	// Test adding duplicate strategy
	err = engine.AddStrategy(strategy)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Test removing strategy
	err = engine.RemoveStrategy("test-strategy-001")
	require.NoError(t, err)
	assert.Len(t, engine.Strategies, 0)

	// Test removing non-existent strategy
	err = engine.RemoveStrategy("non-existent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStrategyEngine_StartStop(t *testing.T) {
	engine := NewStrategyEngine()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test starting the engine
	err := engine.Start(ctx)
	require.NoError(t, err)
	assert.True(t, engine.IsRunning)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Test stopping the engine
	engine.Stop()
	assert.False(t, engine.IsRunning)

	// Test starting already running engine - should not error since we stopped it
	engine.IsRunning = true
	err = engine.Start(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestArbitrageStrategy(t *testing.T) {
	strategy := ArbitrageStrategy()

	assert.Equal(t, "arbitrage_eth_usdc", strategy.ID)
	assert.Equal(t, "ETH-USDC Cross-DEX Arbitrage", strategy.Name)
	assert.Equal(t, StrategyArbitrage, strategy.Type)
	assert.True(t, strategy.IsActive)
	assert.Len(t, strategy.Conditions, 1)
	assert.Len(t, strategy.Actions, 1)

	// Check parameters
	assert.Equal(t, 0.05, strategy.Parameters.MaxPositionSize)
	assert.Equal(t, 0.01, strategy.Parameters.MinProfitMargin)
	assert.Equal(t, 0.005, strategy.Parameters.MaxSlippage)
	assert.Equal(t, 5, strategy.Parameters.ExecutionDelay)
	assert.Equal(t, 60, strategy.Parameters.CooldownPeriod)
	assert.Equal(t, []string{"ETH", "USDC"}, strategy.Parameters.TargetAssets)

	// Check condition
	condition := strategy.Conditions[0]
	assert.Equal(t, "price_diff_condition", condition.ID)
	assert.Equal(t, "price_difference", condition.Metric)
	assert.Equal(t, ">", condition.Operator)
	assert.Equal(t, 0.01, condition.Threshold)
	assert.Equal(t, 30, condition.Duration)
	assert.Equal(t, "average", condition.Aggregation)

	// Check action
	action := strategy.Actions[0]
	assert.Equal(t, "buy_low_sell_high", action.ID)
	assert.Equal(t, ActionSwap, action.Type)
	assert.Equal(t, "uniswap", action.Target)
	assert.Equal(t, 1, action.Priority)
}

func TestYieldFarmingStrategy(t *testing.T) {
	strategy := YieldFarmingStrategy()

	assert.Equal(t, "yield_farming_usdc", strategy.ID)
	assert.Equal(t, "USDC Yield Farming", strategy.Name)
	assert.Equal(t, StrategyYieldFarming, strategy.Type)
	assert.True(t, strategy.IsActive)
	assert.Len(t, strategy.Conditions, 1)
	assert.Len(t, strategy.Actions, 1)

	// Check parameters
	assert.Equal(t, 0.2, strategy.Parameters.MaxPositionSize)
	assert.Equal(t, 0.03, strategy.Parameters.MinProfitMargin)
	assert.Equal(t, 0.01, strategy.Parameters.MaxSlippage)
	assert.Equal(t, 10, strategy.Parameters.ExecutionDelay)
	assert.Equal(t, 3600, strategy.Parameters.CooldownPeriod)
	assert.Equal(t, []string{"USDC"}, strategy.Parameters.TargetAssets)

	// Check condition
	condition := strategy.Conditions[0]
	assert.Equal(t, "yield_condition", condition.ID)
	assert.Equal(t, "yield_rate", condition.Metric)
	assert.Equal(t, ">", condition.Operator)
	assert.Equal(t, 0.05, condition.Threshold)
	assert.Equal(t, 300, condition.Duration)
	assert.Equal(t, "average", condition.Aggregation)

	// Check action
	action := strategy.Actions[0]
	assert.Equal(t, "deposit_usdc", action.ID)
	assert.Equal(t, ActionDeposit, action.Type)
	assert.Equal(t, "aave", action.Target)
	assert.Equal(t, 1, action.Priority)
}

func TestCompareValuesHelper(t *testing.T) {
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
		result := compareValues(tc.a, tc.operator, tc.b)
		assert.Equal(t, tc.expected, result, "Failed for %f %s %f", tc.a, tc.operator, tc.b)
	}
}

func TestActionTypes(t *testing.T) {
	assert.Equal(t, ActionType("swap"), ActionSwap)
	assert.Equal(t, ActionType("deposit"), ActionDeposit)
	assert.Equal(t, ActionType("withdraw"), ActionWithdraw)
	assert.Equal(t, ActionType("borrow"), ActionBorrow)
	assert.Equal(t, ActionType("repay"), ActionRepay)
	assert.Equal(t, ActionType("provide_liquidity"), ActionProvideLiquidity)
	assert.Equal(t, ActionType("remove_liquidity"), ActionRemoveLiquidity)
}

func TestStrategyParametersDefaults(t *testing.T) {
	params := StrategyParameters{
		MaxPositionSize: 0.1,
		MinProfitMargin: 0.01,
		MaxSlippage:     0.005,
		ExecutionDelay:  5,
		CooldownPeriod:  60,
		TargetAssets:    []string{"ETH", "USDC"},
	}

	assert.Equal(t, 0.1, params.MaxPositionSize)
	assert.Equal(t, 0.01, params.MinProfitMargin)
	assert.Equal(t, 0.005, params.MaxSlippage)
	assert.Equal(t, 5, params.ExecutionDelay)
	assert.Equal(t, 60, params.CooldownPeriod)
	assert.Equal(t, []string{"ETH", "USDC"}, params.TargetAssets)
}
