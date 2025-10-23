package defi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// StrategyEngine manages multiple trading strategies
type StrategyEngine struct {
	Strategies map[string]*TradingStrategy
	Agents     map[string]*DeFiAgent
	IsRunning  bool
}

// TradingStrategy defines a complete trading strategy
type TradingStrategy struct {
	ID          string
	Name        string
	Type        StrategyType
	Description string
	Parameters  StrategyParameters
	Conditions  []ExecutionCondition
	Actions     []StrategyAction
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// StrategyParameters contains configurable strategy parameters
type StrategyParameters struct {
	MaxPositionSize float64  `json:"max_position_size"`
	MinProfitMargin float64  `json:"min_profit_margin"`
	MaxSlippage     float64  `json:"max_slippage"`
	ExecutionDelay  int      `json:"execution_delay"` // seconds
	CooldownPeriod  int      `json:"cooldown_period"` // seconds
	TargetAssets    []string `json:"target_assets"`
}

// ExecutionCondition defines when to execute the strategy
type ExecutionCondition struct {
	ID          string                 `json:"id"`
	Metric      string                 `json:"metric"`
	Operator    string                 `json:"operator"`
	Threshold   float64                `json:"threshold"`
	Duration    int                    `json:"duration"`    // seconds
	Aggregation string                 `json:"aggregation"` // "average", "max", "min"
	Metadata    map[string]interface{} `json:"metadata"`
}

// StrategyAction defines what to do when conditions are met
type StrategyAction struct {
	ID         string                 `json:"id"`
	Type       ActionType             `json:"type"`
	Target     string                 `json:"target"` // protocol, token, etc.
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type ActionType string

const (
	ActionSwap             ActionType = "swap"
	ActionDeposit          ActionType = "deposit"
	ActionWithdraw         ActionType = "withdraw"
	ActionBorrow           ActionType = "borrow"
	ActionRepay            ActionType = "repay"
	ActionProvideLiquidity ActionType = "provide_liquidity"
	ActionRemoveLiquidity  ActionType = "remove_liquidity"
)

// NewStrategyEngine creates a new strategy engine
func NewStrategyEngine() *StrategyEngine {
	return &StrategyEngine{
		Strategies: make(map[string]*TradingStrategy),
		Agents:     make(map[string]*DeFiAgent),
		IsRunning:  false,
	}
}

// AddStrategy adds a new trading strategy
func (se *StrategyEngine) AddStrategy(strategy *TradingStrategy) error {
	if _, exists := se.Strategies[strategy.ID]; exists {
		return fmt.Errorf("strategy with ID %s already exists", strategy.ID)
	}

	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = time.Now()
	se.Strategies[strategy.ID] = strategy

	log.Printf("Added strategy: %s (%s)", strategy.Name, strategy.ID)
	return nil
}

// RemoveStrategy removes a strategy
func (se *StrategyEngine) RemoveStrategy(strategyID string) error {
	if _, exists := se.Strategies[strategyID]; !exists {
		return fmt.Errorf("strategy %s not found", strategyID)
	}

	delete(se.Strategies, strategyID)
	log.Printf("Removed strategy: %s", strategyID)
	return nil
}

// Start begins strategy execution
func (se *StrategyEngine) Start(ctx context.Context) error {
	if se.IsRunning {
		return fmt.Errorf("strategy engine is already running")
	}

	se.IsRunning = true
	log.Printf("Strategy engine started")

	// Start strategy evaluation loop
	go se.evaluateStrategies(ctx)

	return nil
}

// Stop halts strategy execution
func (se *StrategyEngine) Stop() {
	se.IsRunning = false
	log.Printf("Strategy engine stopped")
}

// evaluateStrategies continuously evaluates all active strategies
func (se *StrategyEngine) evaluateStrategies(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Evaluate every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !se.IsRunning {
				return
			}

			for _, strategy := range se.Strategies {
				if strategy.IsActive {
					go se.evaluateStrategy(strategy)
				}
			}
		}
	}
}

// evaluateStrategy evaluates a single strategy
func (se *StrategyEngine) evaluateStrategy(strategy *TradingStrategy) {
	// Check if all conditions are met
	conditionsMet := true
	for _, condition := range strategy.Conditions {
		if !se.evaluateCondition(condition) {
			conditionsMet = false
			break
		}
	}

	if conditionsMet {
		log.Printf("Strategy %s conditions met, executing actions", strategy.Name)
		se.executeStrategyActions(strategy)
	}
}

// evaluateCondition evaluates a single condition
func (se *StrategyEngine) evaluateCondition(condition ExecutionCondition) bool {
	currentValue := se.getCurrentMetricValue(condition.Metric, condition.Metadata)
	return compareValues(currentValue, condition.Operator, condition.Threshold)
}

// getCurrentMetricValue retrieves current value for a metric
func (se *StrategyEngine) getCurrentMetricValue(metric string, metadata map[string]interface{}) float64 {
	switch metric {
	case "price_difference":
		return se.calculatePriceDifference(metadata)
	case "yield_rate":
		return se.getCurrentYieldRate(metadata)
	case "liquidity_depth":
		return se.getLiquidityDepth(metadata)
	case "volatility":
		return se.calculateVolatility(metadata)
	case "gas_price":
		return se.getCurrentGasPrice()
	default:
		log.Printf("Unknown metric: %s", metric)
		return 0.0
	}
}

// executeStrategyActions executes all actions for a strategy
func (se *StrategyEngine) executeStrategyActions(strategy *TradingStrategy) {
	// Sort actions by priority
	actions := se.sortActionsByPriority(strategy.Actions)

	for _, action := range actions {
		se.executeAction(action)
	}

	strategy.UpdatedAt = time.Now()
}

// sortActionsByPriority sorts actions by priority (higher first)
func (se *StrategyEngine) sortActionsByPriority(actions []StrategyAction) []StrategyAction {
	sorted := make([]StrategyAction, len(actions))
	copy(sorted, actions)

	// Simple bubble sort for priority (higher numbers first)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j].Priority < sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// executeAction executes a single strategy action
func (se *StrategyEngine) executeAction(action StrategyAction) {
	log.Printf("Executing action: %s (%s)", action.Type, action.ID)

	switch action.Type {
	case ActionSwap:
		se.executeSwapAction(action)
	case ActionDeposit:
		se.executeDepositAction(action)
	case ActionWithdraw:
		se.executeWithdrawAction(action)
	case ActionBorrow:
		se.executeBorrowAction(action)
	case ActionRepay:
		se.executeRepayAction(action)
	case ActionProvideLiquidity:
		se.executeProvideLiquidityAction(action)
	case ActionRemoveLiquidity:
		se.executeRemoveLiquidityAction(action)
	default:
		log.Printf("Unknown action type: %s", action.Type)
	}
}

// Example strategy implementations

// ArbitrageStrategy creates a cross-DEX arbitrage strategy
func ArbitrageStrategy() *TradingStrategy {
	return &TradingStrategy{
		ID:          "arbitrage_eth_usdc",
		Name:        "ETH-USDC Cross-DEX Arbitrage",
		Type:        StrategyArbitrage,
		Description: "Arbitrage between Uniswap and Sushiswap for ETH-USDC pairs",
		Parameters: StrategyParameters{
			MaxPositionSize: 0.05,  // 5% of portfolio
			MinProfitMargin: 0.01,  // 1%
			MaxSlippage:     0.005, // 0.5%
			ExecutionDelay:  5,     // 5 seconds
			CooldownPeriod:  60,    // 1 minute
			TargetAssets:    []string{"ETH", "USDC"},
		},
		Conditions: []ExecutionCondition{
			{
				ID:          "price_diff_condition",
				Metric:      "price_difference",
				Operator:    ">",
				Threshold:   0.01, // 1% difference
				Duration:    30,   // 30 seconds
				Aggregation: "average",
				Metadata: map[string]interface{}{
					"token_pair": "ETH/USDC",
					"dex1":       "uniswap",
					"dex2":       "sushiswap",
				},
			},
		},
		Actions: []StrategyAction{
			{
				ID:     "buy_low_sell_high",
				Type:   ActionSwap,
				Target: "uniswap",
				Parameters: map[string]interface{}{
					"from_token": "USDC",
					"to_token":   "ETH",
					"amount":     1000.0,
				},
				Priority: 1,
			},
		},
		IsActive: true,
	}
}

// YieldFarmingStrategy creates a yield farming strategy
func YieldFarmingStrategy() *TradingStrategy {
	return &TradingStrategy{
		ID:          "yield_farming_usdc",
		Name:        "USDC Yield Farming",
		Type:        StrategyYieldFarming,
		Description: "Deposit USDC to Aave for yield farming",
		Parameters: StrategyParameters{
			MaxPositionSize: 0.2,  // 20% of portfolio
			MinProfitMargin: 0.03, // 3% APY
			MaxSlippage:     0.01, // 1%
			ExecutionDelay:  10,
			CooldownPeriod:  3600, // 1 hour
			TargetAssets:    []string{"USDC"},
		},
		Conditions: []ExecutionCondition{
			{
				ID:          "yield_condition",
				Metric:      "yield_rate",
				Operator:    ">",
				Threshold:   0.05, // 5% APY
				Duration:    300,  // 5 minutes
				Aggregation: "average",
				Metadata: map[string]interface{}{
					"protocol": "aave",
					"asset":    "USDC",
				},
			},
		},
		Actions: []StrategyAction{
			{
				ID:     "deposit_usdc",
				Type:   ActionDeposit,
				Target: "aave",
				Parameters: map[string]interface{}{
					"asset":  "USDC",
					"amount": 5000.0,
				},
				Priority: 1,
			},
		},
		IsActive: true,
	}
}

// Helper functions for metric calculations

func (se *StrategyEngine) calculatePriceDifference(metadata map[string]interface{}) float64 {
	// Implementation would fetch prices from multiple DEXs
	// and calculate the difference
	return 0.015 // 1.5% difference (example)
}

func (se *StrategyEngine) getCurrentYieldRate(metadata map[string]interface{}) float64 {
	// Implementation would fetch current yield rates from protocols
	return 0.067 // 6.7% APY (example)
}

func (se *StrategyEngine) getLiquidityDepth(metadata map[string]interface{}) float64 {
	// Implementation would calculate liquidity depth
	return 5000000.0 // $5M liquidity (example)
}

func (se *StrategyEngine) calculateVolatility(metadata map[string]interface{}) float64 {
	// Implementation would calculate price volatility
	return 0.025 // 2.5% volatility (example)
}

func (se *StrategyEngine) getCurrentGasPrice() float64 {
	// Implementation would fetch current gas price
	return 25.0 // 25 Gwei (example)
}

// Action execution methods (stubs for now)

func (se *StrategyEngine) executeSwapAction(action StrategyAction) {
	log.Printf("Executing swap action: %v", action.Parameters)
	// Implementation would call Uniswap/Sushiswap contracts
}

func (se *StrategyEngine) executeDepositAction(action StrategyAction) {
	log.Printf("Executing deposit action: %v", action.Parameters)
	// Implementation would call Aave/Compound contracts
}

func (se *StrategyEngine) executeWithdrawAction(action StrategyAction) {
	log.Printf("Executing withdraw action: %v", action.Parameters)
	// Implementation would call lending protocol contracts
}

func (se *StrategyEngine) executeBorrowAction(action StrategyAction) {
	log.Printf("Executing borrow action: %v", action.Parameters)
	// Implementation would call lending protocol contracts
}

func (se *StrategyEngine) executeRepayAction(action StrategyAction) {
	log.Printf("Executing repay action: %v", action.Parameters)
	// Implementation would call lending protocol contracts
}

func (se *StrategyEngine) executeProvideLiquidityAction(action StrategyAction) {
	log.Printf("Executing provide liquidity action: %v", action.Parameters)
	// Implementation would call DEX contracts
}

func (se *StrategyEngine) executeRemoveLiquidityAction(action StrategyAction) {
	log.Printf("Executing remove liquidity action: %v", action.Parameters)
	// Implementation would call DEX contracts
}

// Helper function for value comparison
func compareValues(a float64, operator string, b float64) bool {
	switch operator {
	case ">":
		return a > b
	case "<":
		return a < b
	case "==":
		return a == b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	default:
		return false
	}
}

// MarshalJSON for TradingStrategy
func (ts *TradingStrategy) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":          ts.ID,
		"name":        ts.Name,
		"type":        ts.Type,
		"description": ts.Description,
		"parameters":  ts.Parameters,
		"conditions":  ts.Conditions,
		"actions":     ts.Actions,
		"is_active":   ts.IsActive,
		"created_at":  ts.CreatedAt,
		"updated_at":  ts.UpdatedAt,
	})
}
