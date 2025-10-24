package portfolio

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/monitoring"
)

// PortfolioManager manages multiple portfolios
type PortfolioManager struct {
	portfolios map[string]*Portfolio
	mu         sync.RWMutex
	logger     logging.Logger
	monitor    *monitoring.Monitor
}

// NewPortfolioManager creates a new portfolio manager
func NewPortfolioManager(logger logging.Logger, monitor *monitoring.Monitor) *PortfolioManager {
	return &PortfolioManager{
		portfolios: make(map[string]*Portfolio),
		logger:     logger,
		monitor:    monitor,
	}
}

// CreatePortfolio creates a new portfolio
func (pm *PortfolioManager) CreatePortfolio(id, name string, riskProfile RiskProfile) (*Portfolio, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.portfolios[id]; exists {
		return nil, fmt.Errorf("portfolio %s already exists", id)
	}

	portfolio := NewPortfolio(id, name, riskProfile, pm.logger, pm.monitor)
	pm.portfolios[id] = portfolio

	pm.logger.Info("Created portfolio",
		logging.WithString("portfolio_id", id),
		logging.WithString("portfolio_name", name),
		logging.WithString("risk_profile", string(riskProfile.Type)),
	)

	return portfolio, nil
}

// GetPortfolio retrieves a portfolio by ID
func (pm *PortfolioManager) GetPortfolio(id string) (*Portfolio, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	portfolio, exists := pm.portfolios[id]
	if !exists {
		return nil, fmt.Errorf("portfolio %s not found", id)
	}

	return portfolio, nil
}

// DeletePortfolio deletes a portfolio
func (pm *PortfolioManager) DeletePortfolio(id string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.portfolios[id]; !exists {
		return fmt.Errorf("portfolio %s not found", id)
	}

	delete(pm.portfolios, id)
	pm.logger.Info("Deleted portfolio", logging.WithString("portfolio_id", id))

	return nil
}

// ListPortfolios returns all portfolios
func (pm *PortfolioManager) ListPortfolios() []*Portfolio {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	portfolios := make([]*Portfolio, 0, len(pm.portfolios))
	for _, portfolio := range pm.portfolios {
		portfolios = append(portfolios, portfolio)
	}

	return portfolios
}

// GetTotalValue calculates the total value across all portfolios
func (pm *PortfolioManager) GetTotalValue() *big.Float {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	total := big.NewFloat(0)
	for _, portfolio := range pm.portfolios {
		total.Add(total, portfolio.GetTotalValue())
	}

	return total
}

// RebalancePortfolio rebalances a portfolio to target allocations
func (pm *PortfolioManager) RebalancePortfolio(ctx context.Context, portfolioID string) error {
	portfolio, err := pm.GetPortfolio(portfolioID)
	if err != nil {
		return err
	}

	currentAllocation := portfolio.GetAssetAllocation()
	targetAllocation := portfolio.RiskProfile.TargetAllocations

	// Calculate rebalancing actions
	rebalanceActions := calculateRebalanceActions(currentAllocation, targetAllocation, portfolio.GetTotalValue())

	// Execute rebalancing actions
	for _, action := range rebalanceActions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := pm.executeRebalanceAction(ctx, portfolio, action); err != nil {
				pm.logger.Error("Failed to execute rebalance action",
					logging.WithString("portfolio", portfolioID),
					logging.WithString("asset", action.Asset),
					logging.WithError(err),
				)
				return err
			}
		}
	}

	portfolio.LastRebalance = time.Now()
	pm.logger.Info("Portfolio rebalanced",
		logging.WithString("portfolio", portfolioID),
		logging.WithInt("actions_executed", len(rebalanceActions)),
	)

	return nil
}

// RebalanceAction represents a rebalancing action
type RebalanceAction struct {
	Asset  string
	Action RebalanceActionType
	Amount *big.Float
	Reason string
}

type RebalanceActionType string

const (
	ActionBuy    RebalanceActionType = "buy"
	ActionSell   RebalanceActionType = "sell"
	ActionHold   RebalanceActionType = "hold"
	ActionAdjust RebalanceActionType = "adjust"
)

// calculateRebalanceActions calculates the rebalancing actions needed
func calculateRebalanceActions(current, target map[string]float64, totalValue *big.Float) []RebalanceAction {
	var actions []RebalanceAction
	threshold := 2.0 // 2% threshold for rebalancing

	totalFloat, _ := totalValue.Float64()

	for asset, targetPercent := range target {
		currentPercent := current[asset]
		difference := currentPercent - targetPercent

		if abs(difference) > threshold {
			targetValue := (targetPercent / 100) * totalFloat
			currentValue := (currentPercent / 100) * totalFloat
			adjustment := targetValue - currentValue

			action := RebalanceAction{
				Asset:  asset,
				Amount: big.NewFloat(abs(adjustment)),
				Reason: fmt.Sprintf("Current allocation %.2f%% vs target %.2f%%", currentPercent, targetPercent),
			}

			if adjustment > 0 {
				action.Action = ActionBuy
			} else {
				action.Action = ActionSell
			}

			actions = append(actions, action)
		}
	}

	return actions
}

// executeRebalanceAction executes a single rebalancing action
func (pm *PortfolioManager) executeRebalanceAction(ctx context.Context, portfolio *Portfolio, action RebalanceAction) error {
	// In a real implementation, this would execute actual trades
	// For now, we'll just log the action

	amountFloat, _ := action.Amount.Float64()
	pm.logger.Info("Executing rebalance action",
		logging.WithString("portfolio", portfolio.ID),
		logging.WithString("asset", action.Asset),
		logging.WithString("action", string(action.Action)),
		logging.WithFloat64("amount", amountFloat),
		logging.WithString("reason", action.Reason),
	)

	// Simulate trade execution
	time.Sleep(100 * time.Millisecond)

	return nil
}

// abs returns the absolute value of a float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// RiskManager manages portfolio risk
func (pm *PortfolioManager) RiskManager(portfolioID string) (*RiskAssessment, error) {
	portfolio, err := pm.GetPortfolio(portfolioID)
	if err != nil {
		return nil, err
	}

	return pm.assessPortfolioRisk(portfolio), nil
}

// RiskAssessment contains portfolio risk analysis
type RiskAssessment struct {
	PortfolioID     string
	TotalRiskScore  float64
	MaxDrawdown     float64
	Volatility      float64
	SharpeRatio     float64
	Concentration   float64
	RiskLevel       RiskLevel
	Recommendations []string
	AssessmentTime  time.Time
}

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// assessPortfolioRisk assesses the risk of a portfolio
func (pm *PortfolioManager) assessPortfolioRisk(portfolio *Portfolio) *RiskAssessment {
	stats := portfolio.GetPortfolioStats()

	// Calculate concentration risk
	concentration := calculateConcentrationRisk(stats.AssetAllocation)

	// Calculate overall risk score (simplified)
	riskScore := concentration * 0.6 // 60% weight to concentration

	// Add position-based risk
	openPositions := portfolio.GetOpenPositions()
	for _, position := range openPositions {
		if position.Leverage > 1 {
			riskScore += 0.1 // Additional risk for leveraged positions
		}
	}

	// Determine risk level
	var riskLevel RiskLevel
	switch {
	case riskScore < 0.3:
		riskLevel = RiskLevelLow
	case riskScore < 0.6:
		riskLevel = RiskLevelMedium
	case riskScore < 0.8:
		riskLevel = RiskLevelHigh
	default:
		riskLevel = RiskLevelCritical
	}

	// Generate recommendations
	recommendations := generateRiskRecommendations(riskLevel, stats.AssetAllocation, openPositions)

	return &RiskAssessment{
		PortfolioID:     portfolio.ID,
		TotalRiskScore:  riskScore,
		Concentration:   concentration,
		RiskLevel:       riskLevel,
		Recommendations: recommendations,
		AssessmentTime:  time.Now(),
	}
}

// calculateConcentrationRisk calculates portfolio concentration risk
func calculateConcentrationRisk(allocation map[string]float64) float64 {
	if len(allocation) == 0 {
		return 0
	}

	// Calculate Herfindahl-Hirschman Index (HHI)
	hhi := 0.0
	for _, percent := range allocation {
		hhi += (percent / 100) * (percent / 100)
	}

	// Normalize to 0-1 scale
	return hhi
}

// generateRiskRecommendations generates risk mitigation recommendations
func generateRiskRecommendations(riskLevel RiskLevel, allocation map[string]float64, positions []*Position) []string {
	var recommendations []string

	switch riskLevel {
	case RiskLevelCritical:
		recommendations = append(recommendations,
			"Immediate action required: Reduce position sizes",
			"Consider hedging strategies",
			"Diversify across more assets",
		)
	case RiskLevelHigh:
		recommendations = append(recommendations,
			"Consider reducing exposure to largest positions",
			"Monitor positions closely",
			"Set tighter stop-losses",
		)
	case RiskLevelMedium:
		recommendations = append(recommendations,
			"Portfolio risk is acceptable",
			"Continue monitoring allocations",
		)
	case RiskLevelLow:
		recommendations = append(recommendations,
			"Portfolio is well-diversified",
			"Risk level is optimal",
		)
	}

	// Add specific recommendations based on concentration
	if len(allocation) < 3 {
		recommendations = append(recommendations, "Consider adding more assets for better diversification")
	}

	// Add leverage-specific recommendations
	for _, position := range positions {
		if position.Leverage > 3 {
			recommendations = append(recommendations, "Reduce leverage on high-leverage positions")
			break
		}
	}

	return recommendations
}
