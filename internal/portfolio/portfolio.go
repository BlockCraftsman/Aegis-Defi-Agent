package portfolio

import (
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/monitoring"
)

// Asset represents a portfolio asset
type Asset struct {
	Symbol     string
	Name       string
	Amount     *big.Float
	ValueUSD   *big.Float
	PriceUSD   *big.Float
	Allocation float64
	APY        float64
	RiskScore  float64
	LastUpdate time.Time
}

// Position represents a trading position
type Position struct {
	ID           string
	Asset        string
	Type         PositionType
	Size         *big.Float
	EntryPrice   *big.Float
	CurrentPrice *big.Float
	Pnl          *big.Float
	PnlPercent   float64
	Leverage     float64
	Status       PositionStatus
	OpenedAt     time.Time
	ClosedAt     *time.Time
}

type PositionType string

const (
	PositionLong  PositionType = "long"
	PositionShort PositionType = "short"
)

type PositionStatus string

const (
	PositionOpen    PositionStatus = "open"
	PositionClosed  PositionStatus = "closed"
	PositionPending PositionStatus = "pending"
)

// Portfolio represents a trading portfolio
type Portfolio struct {
	ID            string
	Name          string
	TotalValue    *big.Float
	CashBalance   *big.Float
	Assets        map[string]*Asset
	Positions     map[string]*Position
	RiskProfile   RiskProfile
	RebalanceAt   time.Time
	LastRebalance time.Time
	mu            sync.RWMutex
	logger        logging.Logger
	monitor       *monitoring.Monitor
}

// RiskProfile defines portfolio risk tolerance
type RiskProfile struct {
	Type              RiskType
	MaxDrawdown       float64
	MaxPositionSize   float64
	MaxLeverage       float64
	TargetAllocations map[string]float64
}

type RiskType string

const (
	RiskConservative RiskType = "conservative"
	RiskModerate     RiskType = "moderate"
	RiskAggressive   RiskType = "aggressive"
)

// NewPortfolio creates a new portfolio
func NewPortfolio(id, name string, riskProfile RiskProfile, logger logging.Logger, monitor *monitoring.Monitor) *Portfolio {
	return &Portfolio{
		ID:          id,
		Name:        name,
		TotalValue:  big.NewFloat(0),
		CashBalance: big.NewFloat(0),
		Assets:      make(map[string]*Asset),
		Positions:   make(map[string]*Position),
		RiskProfile: riskProfile,
		logger:      logger,
		monitor:     monitor,
	}
}

// AddAsset adds an asset to the portfolio
func (p *Portfolio) AddAsset(asset *Asset) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.Assets[asset.Symbol]; exists {
		return fmt.Errorf("asset %s already exists in portfolio", asset.Symbol)
	}

	p.Assets[asset.Symbol] = asset
	amountFloat, _ := asset.Amount.Float64()
	p.logger.Info("Added asset to portfolio",
		logging.WithString("portfolio", p.ID),
		logging.WithString("asset", asset.Symbol),
		logging.WithFloat64("amount", amountFloat),
	)

	if p.monitor != nil {
		p.monitor.RecordPortfolioUpdate(p.ID, len(p.Assets), len(p.Positions))
	}

	return nil
}

// UpdateAsset updates an asset in the portfolio
func (p *Portfolio) UpdateAsset(symbol string, amount, priceUSD *big.Float) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	asset, exists := p.Assets[symbol]
	if !exists {
		return fmt.Errorf("asset %s not found in portfolio", symbol)
	}

	asset.Amount = amount
	asset.PriceUSD = priceUSD
	asset.ValueUSD = new(big.Float).Mul(amount, priceUSD)
	asset.LastUpdate = time.Now()

	amountFloat, _ := amount.Float64()
	priceFloat, _ := priceUSD.Float64()
	p.logger.Debug("Updated asset",
		logging.WithString("portfolio", p.ID),
		logging.WithString("asset", symbol),
		logging.WithFloat64("amount", amountFloat),
		logging.WithFloat64("price", priceFloat),
	)

	return nil
}

// RemoveAsset removes an asset from the portfolio
func (p *Portfolio) RemoveAsset(symbol string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.Assets[symbol]; !exists {
		return fmt.Errorf("asset %s not found in portfolio", symbol)
	}

	delete(p.Assets, symbol)
	p.logger.Info("Removed asset from portfolio",
		logging.WithString("portfolio", p.ID),
		logging.WithString("asset", symbol),
	)

	if p.monitor != nil {
		p.monitor.RecordPortfolioUpdate(p.ID, len(p.Assets), len(p.Positions))
	}

	return nil
}

// OpenPosition opens a new trading position
func (p *Portfolio) OpenPosition(position *Position) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.Positions[position.ID]; exists {
		return fmt.Errorf("position %s already exists", position.ID)
	}

	position.Status = PositionOpen
	position.OpenedAt = time.Now()
	p.Positions[position.ID] = position

	sizeFloat, _ := position.Size.Float64()
	p.logger.Info("Opened position",
		logging.WithString("portfolio", p.ID),
		logging.WithString("position", position.ID),
		logging.WithString("asset", position.Asset),
		logging.WithString("type", string(position.Type)),
		logging.WithFloat64("size", sizeFloat),
	)

	if p.monitor != nil {
		p.monitor.RecordPortfolioUpdate(p.ID, len(p.Assets), len(p.Positions))
		p.monitor.RecordPositionOpened(position.Asset, string(position.Type))
	}

	return nil
}

// ClosePosition closes an existing position
func (p *Portfolio) ClosePosition(positionID string, exitPrice *big.Float) (*Position, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	position, exists := p.Positions[positionID]
	if !exists {
		return nil, fmt.Errorf("position %s not found", positionID)
	}

	if position.Status != PositionOpen {
		return nil, fmt.Errorf("position %s is not open", positionID)
	}

	position.CurrentPrice = exitPrice
	position.Status = PositionClosed
	now := time.Now()
	position.ClosedAt = &now

	// Calculate P&L
	if position.Type == PositionLong {
		position.Pnl = new(big.Float).Sub(
			new(big.Float).Mul(position.Size, exitPrice),
			new(big.Float).Mul(position.Size, position.EntryPrice),
		)
	} else {
		position.Pnl = new(big.Float).Sub(
			new(big.Float).Mul(position.Size, position.EntryPrice),
			new(big.Float).Mul(position.Size, exitPrice),
		)
	}

	entryValue := new(big.Float).Mul(position.Size, position.EntryPrice)
	if entryValue.Cmp(big.NewFloat(0)) != 0 {
		pnlFloat, _ := position.Pnl.Float64()
		entryFloat, _ := entryValue.Float64()
		position.PnlPercent = (pnlFloat / entryFloat) * 100
	}

	pnlFloat, _ := position.Pnl.Float64()
	p.logger.Info("Closed position",
		logging.WithString("portfolio", p.ID),
		logging.WithString("position", position.ID),
		logging.WithFloat64("pnl", pnlFloat),
		logging.WithFloat64("pnl_percent", position.PnlPercent),
	)

	if p.monitor != nil {
		p.monitor.RecordPortfolioUpdate(p.ID, len(p.Assets), len(p.Positions))
		p.monitor.RecordPositionClosed(position.Asset, string(position.Type), position.Pnl)
	}

	return position, nil
}

// GetTotalValue calculates the total portfolio value
func (p *Portfolio) GetTotalValue() *big.Float {
	p.mu.RLock()
	defer p.mu.RUnlock()

	total := new(big.Float).Set(p.CashBalance)

	for _, asset := range p.Assets {
		if asset.ValueUSD != nil {
			total.Add(total, asset.ValueUSD)
		}
	}

	return total
}

// GetAssetAllocation calculates asset allocation percentages
func (p *Portfolio) GetAssetAllocation() map[string]float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalValue := p.GetTotalValue()
	allocation := make(map[string]float64)

	for symbol, asset := range p.Assets {
		if asset.ValueUSD != nil && totalValue.Cmp(big.NewFloat(0)) != 0 {
			valueFloat, _ := asset.ValueUSD.Float64()
			totalFloat, _ := totalValue.Float64()
			allocation[symbol] = (valueFloat / totalFloat) * 100
		}
	}

	return allocation
}

// GetOpenPositions returns all open positions
func (p *Portfolio) GetOpenPositions() []*Position {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var openPositions []*Position
	for _, position := range p.Positions {
		if position.Status == PositionOpen {
			openPositions = append(openPositions, position)
		}
	}

	return openPositions
}

// GetPortfolioStats returns portfolio statistics
func (p *Portfolio) GetPortfolioStats() *PortfolioStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &PortfolioStats{
		TotalValue:      p.GetTotalValue(),
		CashBalance:     new(big.Float).Set(p.CashBalance),
		AssetCount:      len(p.Assets),
		OpenPositions:   len(p.GetOpenPositions()),
		TotalPositions:  len(p.Positions),
		AssetAllocation: p.GetAssetAllocation(),
		LastUpdate:      time.Now(),
	}

	return stats
}

// PortfolioStats contains portfolio statistics
type PortfolioStats struct {
	TotalValue      *big.Float
	CashBalance     *big.Float
	AssetCount      int
	OpenPositions   int
	TotalPositions  int
	AssetAllocation map[string]float64
	LastUpdate      time.Time
}
