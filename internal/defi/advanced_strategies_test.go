package defi

import (
	"context"
	"testing"
)

func TestAdvancedStrategyEngine(t *testing.T) {
	engine := NewAdvancedStrategyEngine()

	// Test engine initialization
	if engine == nil {
		t.Fatal("Failed to create advanced strategy engine")
	}

	if engine.Strategies == nil {
		t.Error("Strategies map not initialized")
	}

	if engine.RiskManager == nil {
		t.Error("Risk manager not initialized")
	}

	if engine.MarketAnalyzer == nil {
		t.Error("Market analyzer not initialized")
	}

	if engine.Portfolio == nil {
		t.Error("Portfolio manager not initialized")
	}
}

func TestAdvancedStrategiesCreation(t *testing.T) {
	// Test mean reversion strategy
	meanReversion := MeanReversionStrategy()
	if meanReversion == nil {
		t.Fatal("Failed to create mean reversion strategy")
	}

	if meanReversion.ID != "mean_reversion_eth" {
		t.Errorf("Expected ID 'mean_reversion_eth', got '%s'", meanReversion.ID)
	}

	if meanReversion.Type != StrategyArbitrage {
		t.Errorf("Expected strategy type StrategyArbitrage, got %v", meanReversion.Type)
	}

	// Test trend following strategy
	trendFollowing := TrendFollowingStrategy()
	if trendFollowing == nil {
		t.Fatal("Failed to create trend following strategy")
	}

	if trendFollowing.ID != "trend_following_btc" {
		t.Errorf("Expected ID 'trend_following_btc', got '%s'", trendFollowing.ID)
	}

	// Test statistical arbitrage strategy
	statArb := StatisticalArbitrageStrategy()
	if statArb == nil {
		t.Fatal("Failed to create statistical arbitrage strategy")
	}

	if statArb.ID != "stat_arb_eth_btc" {
		t.Errorf("Expected ID 'stat_arb_eth_btc', got '%s'", statArb.ID)
	}
}

func TestPositionSizing(t *testing.T) {
	engine := NewAdvancedStrategyEngine()
	strategy := MeanReversionStrategy()
	portfolioValue := 100000.0 // $100k portfolio

	// Test Kelly criterion position sizing
	positionSize := engine.CalculatePositionSize(strategy, portfolioValue)

	// Position size should be reasonable (between 0.5% and 10% of portfolio)
	if positionSize < 500 || positionSize > 10000 {
		t.Errorf("Position size %.2f outside reasonable range (500-10000)", positionSize)
	}

	// Test with different strategy types
	trendStrategy := TrendFollowingStrategy()
	trendPosition := engine.CalculatePositionSize(trendStrategy, portfolioValue)

	if trendPosition <= 0 {
		t.Error("Trend strategy position size should be positive")
	}

	statArbStrategy := StatisticalArbitrageStrategy()
	statArbPosition := engine.CalculatePositionSize(statArbStrategy, portfolioValue)

	if statArbPosition <= 0 {
		t.Error("Statistical arbitrage position size should be positive")
	}
}

func TestRiskAssessment(t *testing.T) {
	engine := NewAdvancedStrategyEngine()
	strategy := MeanReversionStrategy()

	// Set up initial risk parameters
	engine.RiskManager.MarketRisk.Volatility = 0.02
	engine.RiskManager.PortfolioRisk.MaxDrawdown = 0.03
	engine.RiskManager.PortfolioRisk.ValueAtRisk = 0.02

	// Test risk assessment with acceptable parameters
	riskAcceptable := engine.AssessRisk(strategy)
	if !riskAcceptable {
		t.Error("Risk assessment should pass with acceptable parameters")
	}

	// Test risk assessment with high volatility
	engine.RiskManager.MarketRisk.Volatility = 0.05 // Exceeds limit
	riskHighVol := engine.AssessRisk(strategy)
	if riskHighVol {
		t.Error("Risk assessment should fail with high volatility")
	}

	// Reset and test with high drawdown
	engine.RiskManager.MarketRisk.Volatility = 0.02
	engine.RiskManager.PortfolioRisk.MaxDrawdown = 0.06 // Exceeds limit
	riskHighDrawdown := engine.AssessRisk(strategy)
	if riskHighDrawdown {
		t.Error("Risk assessment should fail with high drawdown")
	}
}

func TestStrategyEngineLifecycle(t *testing.T) {
	engine := NewAdvancedStrategyEngine()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Test starting the engine
	err := engine.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start engine: %v", err)
	}

	if !engine.IsRunning {
		t.Error("Engine should be running after Start()")
	}

	// Test adding strategies
	meanReversion := MeanReversionStrategy()
	engine.Strategies[meanReversion.ID] = meanReversion

	if len(engine.Strategies) != 1 {
		t.Errorf("Expected 1 strategy, got %d", len(engine.Strategies))
	}

	// Test stopping the engine
	engine.Stop()
	if engine.IsRunning {
		t.Error("Engine should not be running after Stop()")
	}
}

func TestConditionScoring(t *testing.T) {
	engine := NewAdvancedStrategyEngine()

	// Test empty conditions
	emptyScore := engine.calculateConditionScore([]AdvancedCondition{})
	if emptyScore != 0.0 {
		t.Errorf("Empty conditions should return 0.0, got %.2f", emptyScore)
	}

	// Test single condition
	conditions := []AdvancedCondition{
		{
			ID:        "test_condition",
			Metric:    "test_metric",
			Operator:  ">",
			Threshold: 1.0,
			Weight:    1.0,
		},
	}

	score := engine.calculateConditionScore(conditions)
	// Since getAdvancedMetricValue returns 0.0 for unknown metrics, score should be 0
	if score != 0.0 {
		t.Errorf("Unknown metric should return 0.0, got %.2f", score)
	}
}

func TestPortfolioValue(t *testing.T) {
	portfolio := NewPortfolioManager()

	// Test empty portfolio
	emptyValue := portfolio.GetTotalValue()
	if emptyValue != 0.0 {
		t.Errorf("Empty portfolio should have value 0.0, got %.2f", emptyValue)
	}

	// Test portfolio with assets
	portfolio.Assets["ETH"] = &PortfolioAsset{
		Symbol: "ETH",
		Amount: 10.0,
		Value:  20000.0, // $20k
	}

	portfolio.Assets["BTC"] = &PortfolioAsset{
		Symbol: "BTC",
		Amount: 0.5,
		Value:  15000.0, // $15k
	}

	totalValue := portfolio.GetTotalValue()
	expectedValue := 35000.0
	if totalValue != expectedValue {
		t.Errorf("Expected portfolio value %.2f, got %.2f", expectedValue, totalValue)
	}
}

func TestAdvancedRiskManager(t *testing.T) {
	riskManager := NewAdvancedRiskManager()

	if riskManager == nil {
		t.Fatal("Failed to create advanced risk manager")
	}

	if riskManager.PortfolioRisk == nil {
		t.Error("Portfolio risk not initialized")
	}

	if riskManager.MarketRisk == nil {
		t.Error("Market risk not initialized")
	}

	if riskManager.LiquidityRisk == nil {
		t.Error("Liquidity risk not initialized")
	}

	if riskManager.CounterpartyRisk == nil {
		t.Error("Counterparty risk not initialized")
	}

	if riskManager.RiskLimits == nil {
		t.Error("Risk limits not initialized")
	}

	// Test default risk limits
	if riskManager.RiskLimits.MaxVaR != 0.05 {
		t.Errorf("Expected MaxVaR 0.05, got %.2f", riskManager.RiskLimits.MaxVaR)
	}

	if riskManager.RiskLimits.MaxDrawdown != 0.10 {
		t.Errorf("Expected MaxDrawdown 0.10, got %.2f", riskManager.RiskLimits.MaxDrawdown)
	}
}
