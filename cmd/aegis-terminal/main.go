package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/defi"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/pkg/mcpclient"
)

// DeFiAgentTerminal represents the terminal interface for DeFi AI Agent
type DeFiAgentTerminal struct {
	textinput.Model
	updateTimer    timer.Model
	defiAgent      *defi.DeFiAgent
	strategyEngine *defi.StrategyEngine
	priceClient    *mcpclient.CoinGeckoClient
	messages       []string
	isRunning      bool
	currentView    string
	marketData     MarketData
	walletData     WalletData
	commandHistory []string
	currentHistory int
	isFullScreen   bool
	tradingData    TradingData
}

// TradingData holds trading chart and operation data
type TradingData struct {
	PriceHistory   []PricePoint
	ActivePair     string
	ChartWidth     int
	ChartHeight    int
	TradeHistory   []Trade
	CurrentBalance float64
}

// PricePoint represents a single price point for charting
type PricePoint struct {
	Timestamp time.Time
	Price     float64
	Volume    float64
}

// Trade represents a trading operation
type Trade struct {
	ID        string
	Pair      string
	Type      string // "buy" or "sell"
	Amount    float64
	Price     float64
	Timestamp time.Time
	Profit    float64
}

// MarketData holds real-time market information
type MarketData struct {
	ETHPrice     float64
	BTCPrice     float64
	USDCPrice    float64
	USDTPrice    float64
	ArbitrageOps []ArbitrageOpportunity
	YieldRates   map[string]float64
	GasPrice     int
}

// WalletData holds portfolio information
type WalletData struct {
	TotalValue   float64
	ETHBalance   float64
	USDCBalance  float64
	RecentProfit float64
}

// ArbitrageOpportunity represents a detected arbitrage opportunity
type ArbitrageOpportunity struct {
	Pair      string
	ExchangeA string
	ExchangeB string
	PriceA    float64
	PriceB    float64
	Profit    float64
	ProfitPct float64
}

// NewDeFiAgentTerminal creates a new terminal interface
func NewDeFiAgentTerminal() DeFiAgentTerminal {
	ti := textinput.New()
	ti.Placeholder = "Enter DeFi command (e.g., 'deploy arbitrage', 'check balance', 'market status')"
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 80

	// Initialize timer for real-time updates
	updateTimer := timer.NewWithInterval(5*time.Second, 5*time.Second)

	// Initialize DeFi components
	strategyEngine := defi.NewStrategyEngine()

	// Add sample strategies
	strategyEngine.AddStrategy(defi.ArbitrageStrategy())
	strategyEngine.AddStrategy(defi.YieldFarmingStrategy())

	// Initialize CoinGecko client for real price data
	priceClient := mcpclient.NewCoinGeckoClient()

	// Initialize market data with zero values - will be populated by real data
	marketData := MarketData{
		ETHPrice:  0.0,
		BTCPrice:  0.0,
		USDCPrice: 0.0,
		USDTPrice: 0.0,
		GasPrice:  28,
		YieldRates: map[string]float64{
			"Aave USDC":    4.8,
			"Compound ETH": 3.2,
			"Uniswap V3":   8.5,
		},
	}

	// Initialize wallet data with realistic values
	walletData := WalletData{
		TotalValue:   21580.25,
		ETHBalance:   2.15,
		USDCBalance:  12500.50,
		RecentProfit: 842.75,
	}

	// Initialize trading data
	tradingData := TradingData{
		ActivePair:     "ETH/USDC",
		ChartWidth:     60,
		ChartHeight:    15,
		CurrentBalance: 10000.0,
		PriceHistory:   []PricePoint{},
		TradeHistory:   []Trade{},
	}

	// Generate initial price history
	for i := 0; i < 50; i++ {
		timeAgo := time.Now().Add(-time.Duration(50-i) * time.Minute)
		basePrice := 3800.0 + rand.Float64()*200
		tradingData.PriceHistory = append(tradingData.PriceHistory, PricePoint{
			Timestamp: timeAgo,
			Price:     basePrice + (rand.Float64()-0.5)*50,
			Volume:    rand.Float64() * 1000,
		})
	}

	terminal := DeFiAgentTerminal{
		Model:          ti,
		updateTimer:    updateTimer,
		strategyEngine: strategyEngine,
		priceClient:    priceClient,
		messages:       []string{},
		isRunning:      true,
		currentView:    "dashboard",
		marketData:     marketData,
		walletData:     walletData,
		commandHistory: []string{},
		currentHistory: -1,
		isFullScreen:   false,
		tradingData:    tradingData,
	}

	// Initialize with real market data
	terminal.updateMarketData()

	return terminal
}

func (m DeFiAgentTerminal) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		m.updateTimer.Init(),
	)
}

func (m DeFiAgentTerminal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			command := m.Value()
			if command != "" {
				m.commandHistory = append(m.commandHistory, command)
				m.currentHistory = -1
			}
			m.messages = append(m.messages, fmt.Sprintf("$ %s", command))
			response := m.processCommand(command)
			m.messages = append(m.messages, response)
			m.SetValue("")
		case "up":
			if len(m.commandHistory) > 0 {
				if m.currentHistory == -1 {
					m.currentHistory = len(m.commandHistory) - 1
				} else if m.currentHistory > 0 {
					m.currentHistory--
				}
				if m.currentHistory >= 0 && m.currentHistory < len(m.commandHistory) {
					m.SetValue(m.commandHistory[m.currentHistory])
				}
			}
		case "down":
			if len(m.commandHistory) > 0 {
				if m.currentHistory < len(m.commandHistory)-1 {
					m.currentHistory++
					m.SetValue(m.commandHistory[m.currentHistory])
				} else {
					m.currentHistory = -1
					m.SetValue("")
				}
			}
		case "tab":
			m.autoCompleteCommand()
		case "esc":
			m.currentView = "dashboard"
		case "1":
			if m.currentView == "trading" {
				m.tradingData.ActivePair = "ETH/USDC"
				m.messages = append(m.messages, "🔄 Switched to ETH/USDC trading pair")
			} else {
				m.currentView = "dashboard"
			}
		case "2":
			if m.currentView == "trading" {
				m.tradingData.ActivePair = "BTC/USD"
				m.messages = append(m.messages, "🔄 Switched to BTC/USD trading pair")
			} else {
				m.currentView = "agents"
			}
		case "3":
			if m.currentView == "trading" {
				m.tradingData.ActivePair = "SOL/USD"
				m.messages = append(m.messages, "🔄 Switched to SOL/USD trading pair")
			} else {
				m.currentView = "market"
			}
		case "4":
			m.currentView = "wallet"
		case "5":
			m.currentView = "strategies"
		case "6":
			m.currentView = "trading"
		case "f":
			m.isFullScreen = !m.isFullScreen
		case "t":
			// Quick trading command
			if m.currentView == "trading" {
				m.executeQuickTrade()
			}
		case "b":
			// Buy command
			if m.currentView == "trading" {
				m.executeBuy(0.1) // Default buy 0.1 ETH
			}
		case "s":
			// Sell command
			if m.currentView == "trading" {
				m.executeSell(0.1) // Default sell 0.1 ETH
			}
		}
	case timer.TickMsg:
		m.updateMarketData()
		m.updateTimer, cmd = m.updateTimer.Update(msg)
		return m, cmd
	}

	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m DeFiAgentTerminal) View() string {
	var content string

	// ASCII Art Header with cyberpunk style
	asciiArt := `
    █████╗ ███████╗ ██████╗ ██╗███████╗    ██████╗ ███████╗███████╗██╗    ██████╗ 
   ██╔══██╗██╔════╝██╔════╝ ██║██╔════╝    ██╔══██╗██╔════╝██╔════╝██║    ██╔══██╗
   ███████║█████╗  ██║  ███╗██║█████╗      ██║  ██║█████╗  █████╗  ██║    ██████╔╝
   ██╔══██║██╔══╝  ██║   ██║██║██╔══╝      ██║  ██║██╔══╝  ██╔══╝  ██║    ██╔══██╗
   ██║  ██║███████╗╚██████╔╝██║███████╗    ██████╔╝███████╗███████╗██║    ██████╔╝
   ╚═╝  ╚═╝╚══════╝ ╚═════╝ ╚═╝╚══════╝    ╚═════╝ ╚══════╝╚══════╝╚═╝    ╚═════╝ 
                                                                                   
                          ███████╗██╗  ██╗██████╗ ███████╗██████╗                  
                          ██╔════╝╚██╗██╔╝██╔══██╗██╔════╝██╔══██╗                 
                          █████╗   ╚███╔╝ ██║  ██║█████╗  ██████╔╝                 
                          ██╔══╝   ██╔██╗ ██║  ██║██╔══╝  ██╔══██╗                 
                          ███████╗██╔╝ ██╗██████╔╝███████╗██║  ██║                 
                          ╚══════╝╚═╝  ╚═╝╚═════╝ ╚══════╝╚═╝  ╚═╝                 
`

	// Header with ASCII art
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	content += headerStyle.Render(asciiArt) + "\n"

	// Status bar
	statusBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Background(lipgloss.Color("#1A1A1A")).
		Bold(true).
		Padding(0, 1).
		Width(80)

	fullScreenIndicator := ""
	if m.isFullScreen {
		fullScreenIndicator = " | FULL SCREEN"
	}
	statusText := fmt.Sprintf("🟢 SYSTEM ACTIVE | VIEW: %s | AGENTS: 4 | TVL: $%.2f%s",
		strings.ToUpper(m.currentView), m.walletData.TotalValue, fullScreenIndicator)
	content += statusBar.Render(statusText) + "\n\n"

	// Navigation with cyberpunk style
	navContainer := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF00FF")).
		Padding(0, 1)

	navStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Padding(0, 1)

	activeNavStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF00FF")).
		Bold(true).
		Background(lipgloss.Color("#1A1A1A")).
		Padding(0, 1)

	navItems := []string{
		"[1] DASHBOARD",
		"[2] AGENTS",
		"[3] MARKET",
		"[4] WALLET",
		"[5] STRATEGIES",
		"[6] TRADING",
	}

	var navDisplay []string
	for i, item := range navItems {
		if fmt.Sprintf("%d", i+1) == m.currentView {
			navDisplay = append(navDisplay, activeNavStyle.Render(item))
		} else {
			navDisplay = append(navDisplay, navStyle.Render(item))
		}
	}

	navContent := lipgloss.JoinHorizontal(lipgloss.Left, navDisplay...)
	content += navContainer.Render(navContent) + "\n\n"

	// Main content with cyberpunk frame
	frameStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#00FF00")).
		Padding(0, 1)

	var mainContent string
	switch m.currentView {
	case "dashboard":
		mainContent = m.renderDashboard()
	case "agents":
		mainContent = m.renderAgents()
	case "market":
		mainContent = m.renderMarket()
	case "wallet":
		mainContent = m.renderWallet()
	case "strategies":
		mainContent = m.renderStrategies()
	case "trading":
		mainContent = m.renderTradingView()
	}

	content += frameStyle.Render(mainContent)

	// Command history with terminal style
	if len(m.messages) > 0 {
		content += "\n\n" + m.renderCommandHistory()
	}

	// Cyberpunk input prompt
	content += "\n\n" + m.renderCyberpunkInput()

	// Footer with glitch effect
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF00FF")).
		Bold(true).
		Italic(true)

	footer := footerStyle.Render("┃ [Q] QUIT • [1-5] NAVIGATE • [ENTER] EXECUTE • [ESC] DASHBOARD ┃")
	content += "\n" + footer

	return content
}

func (m DeFiAgentTerminal) renderDashboard() string {
	totalValue := (m.walletData.ETHBalance * m.marketData.ETHPrice) + m.walletData.USDCBalance
	profitPercentage := (m.walletData.RecentProfit / totalValue) * 100

	return fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────────────────┐
│                            SYSTEM DASHBOARD                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  🟢 SYSTEM STATUS: ACTIVE                                               │
│  📊 ACTIVE AGENTS: 4                                                    │
│  💰 TOTAL VALUE: $%.2f                                                │
│  📈 24H PROFIT: +$%.2f (+%.1f%%)                                      │
│  ⚡ PERFORMANCE: 98.7%%                                                   │
│                                                                         │
│  ┌─────────────────┬─────────────────┬─────────────────┐                │
│  │   ARBITRAGE     │   YIELD FARMING │   RISK MGMT     │                │
│  ├─────────────────┼─────────────────┼─────────────────┤                │
│  │  Active: 2      │  Active: 1      │  Active: 1      │                │
│  │  Profit: $245   │  APY: 6.7%%      │  Score: 92/100  │                │
│  │  Trades: 47     │  Deposits: 3    │  Alerts: 0      │                │
│  └─────────────────┴─────────────────┴─────────────────┘                │
│                                                                         │
│  Quick Commands:                                                        │
│    • 'deploy arbitrage' - Start arbitrage detection                     │
│    • 'check balance' - Show wallet balances                             │
│    • 'market status' - Current market conditions                        │
│    • 'agent status' - Agent performance                                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`,
		totalValue, m.walletData.RecentProfit, profitPercentage)
}

func (m DeFiAgentTerminal) renderAgents() string {
	return `
┌─────────────────────────────────────────────────────────────────────────┐
│                            AI AGENTS STATUS                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  🟢 HEDERA-ARBITRAGE-001                                                │
│     Status: Active | Efficiency: 98.7% | Trades: 47                     │
│     Last Action: ETH/USDC arbitrage +$124.50                            │
│                                                                         │
│  🟢 PYTH-PRICE-BOT-002                                                  │
│     Status: Active | Accuracy: 99.2% | Updates: 1,247                   │
│     Last Action: Price feed sync completed                              │
│                                                                         │
│  🔵 BLOCKSCOUT-SCANNER-003                                              │
│     Status: Scanning | Coverage: 87.3% | Blocks: 18,452,117             │
│     Last Action: Transaction monitoring active                          │
│                                                                         │
│  🟡 LIT-AUTOMATION-004                                                  │
│     Status: Standby | Ready: 100% | Conditions: 12                      │
│     Last Action: Conditional trigger set                                │
│                                                                         │
│  Commands:                                                              │
│    • 'start agent <name>' - Start specific agent                        │
│    • 'stop agent <name>' - Stop specific agent                          │
│    • 'agent performance' - Detailed performance metrics                 │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`
}

func (m DeFiAgentTerminal) renderMarket() string {
	var arbitrageSection string
	if len(m.marketData.ArbitrageOps) > 0 {
		for _, opp := range m.marketData.ArbitrageOps {
			arbitrageSection += fmt.Sprintf("     %s: %s $%.2f vs %s $%.2f (+%.2f%%)\n",
				opp.Pair, opp.ExchangeA, opp.PriceA, opp.ExchangeB, opp.PriceB, opp.ProfitPct)
		}
	} else {
		arbitrageSection = "     No active arbitrage opportunities detected\n"
	}

	return fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────────────────┐
│                         MARKET INTELLIGENCE                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  📊 REAL-TIME PRICES:                                                   │
│     ETH: $%.2f   BTC: $%.2f                                            │
│     USDC: $%.2f    USDT: $%.2f                                         │
│                                                                         │
│  🔍 ARBITRAGE OPPORTUNITIES:                                            │
%s│                                                                         │
│  📈 YIELD OPPORTUNITIES:                                                │
│     Aave USDC: %.1f%% APY | Compound ETH: %.1f%% APY                    │
│     Uniswap V3 ETH/USDC: %.1f%% APY                                     │
│                                                                         │
│  ⚠️  MARKET ALERTS:                                                     │
│     • Gas prices: %d Gwei                                               │
│     • %d active opportunities                                           │
│     • No critical alerts                                                │
│                                                                         │
│  Commands:                                                              │
│    • 'scan arbitrage' - Scan for new opportunities                      │
│    • 'yield rates' - Current yield farming rates                        │
│    • 'gas status' - Current network conditions                          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`,
		m.marketData.ETHPrice, m.marketData.BTCPrice,
		m.marketData.USDCPrice, m.marketData.USDTPrice,
		arbitrageSection,
		m.marketData.YieldRates["Aave USDC"], m.marketData.YieldRates["Compound ETH"],
		m.marketData.YieldRates["Uniswap V3"],
		m.marketData.GasPrice, len(m.marketData.ArbitrageOps))
}

func (m DeFiAgentTerminal) renderWallet() string {
	ethValue := m.walletData.ETHBalance * m.marketData.ETHPrice
	totalValue := ethValue + m.walletData.USDCBalance
	ethPercentage := (ethValue / totalValue) * 100
	usdcPercentage := (m.walletData.USDCBalance / totalValue) * 100

	return fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────────────────┐
│                           WALLET MANAGEMENT                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  💰 BALANCE OVERVIEW:                                                   │
│     Total Value: $%.2f                                                │
│     24h Change: +$%.2f (+%.1f%%)                                      │
│                                                                         │
│  📊 ASSET ALLOCATION:                                                   │
│     ETH: %.2f ($%.2f) - %.1f%%                                        │
│     USDC: %.2f - %.1f%%                                               │
│                                                                         │
│  🔗 CONNECTED NETWORKS:                                                 │
│     Ethereum Mainnet: 0x3C78f7434AF52Cae4CB71a89C3ACab3BAb9513d6       │
│     Polygon: 0x3C78f7434AF52Cae4CB71a89C3ACab3BAb9513d6                │
│     BSC: 0x3C78f7434AF52Cae4CB71a89C3ACab3BAb9513d6                    │
│                                                                         │
│  📋 RECENT TRANSACTIONS:                                                │
│     • ETH/USDC arbitrage +$124.50 (2 min ago)                           │
│     • USDC deposit to Aave +$1,000 (15 min ago)                         │
│     • Gas fee -$8.75 (30 min ago)                                       │
│                                                                         │
│  Commands:                                                              │
│    • 'balance' - Detailed balance breakdown                             │
│    • 'transactions' - Recent transaction history                        │
│    • 'export keys' - Export wallet information                          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`,
		totalValue, m.walletData.RecentProfit, (m.walletData.RecentProfit/totalValue)*100,
		m.walletData.ETHBalance, ethValue, ethPercentage,
		m.walletData.USDCBalance, usdcPercentage)
}

func (m DeFiAgentTerminal) renderStrategies() string {
	return `
┌─────────────────────────────────────────────────────────────────────────┐
│                          TRADING STRATEGIES                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  🤖 ACTIVE STRATEGIES:                                                  │
│                                                                         │
│  🟢 Cross-DEX Arbitrage                                                 │
│     Status: Active | Profit: $245.00 | Trades: 47                       │
│     Min Profit: 1.0% | Max Slippage: 0.5%                              │
│     Pairs: ETH/USDC, AVAX/USDT                                          │
│                                                                         │
│  🟢 USDC Yield Farming                                                  │
│     Status: Active | APY: 6.7% | Deposits: $5,000                       │
│     Protocol: Aave | Asset: USDC                                        │
│     Min Yield: 5.0%                                                     │
│                                                                         │
│  ⚙️  STRATEGY CONFIGURATION:                                            │
│     • Max Position Size: 10% of portfolio                               │
│     • Stop Loss: 2% per trade                                           │
│     • Cooldown Period: 60 seconds                                       │
│     • Risk Score: 92/100                                                │
│                                                                         │
│  Commands:                                                              │
│    • 'create strategy' - Create new trading strategy                    │
│    • 'modify strategy' - Adjust strategy parameters                     │
│    • 'backtest' - Run strategy backtesting                              │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`
}

func (m DeFiAgentTerminal) renderCommandHistory() string {
	var history string
	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Width(80)

	// Show last 5 messages
	start := len(m.messages) - 5
	if start < 0 {
		start = 0
	}

	for i := start; i < len(m.messages); i++ {
		history += messageStyle.Render(m.messages[i]) + "\n"
	}

	return history
}

// getStrategyStatus returns the current status of all strategies
func (m DeFiAgentTerminal) getStrategyStatus() string {
	var status string
	status += "🤖 STRATEGY STATUS:\n"

	for _, strategy := range m.strategyEngine.Strategies {
		statusType := "🟢"
		if !strategy.IsActive {
			statusType = "🔴"
		}

		status += fmt.Sprintf("%s %s - %s\n", statusType, strategy.Name, strategy.Description)
		status += fmt.Sprintf("   Type: %s | Active: %v\n", strategy.Type, strategy.IsActive)
		status += fmt.Sprintf("   Max Position: %.1f%% | Min Profit: %.1f%%\n",
			strategy.Parameters.MaxPositionSize*100, strategy.Parameters.MinProfitMargin*100)
		status += "\n"
	}

	if len(m.strategyEngine.Strategies) == 0 {
		status = "No strategies configured. Use 'create strategy' to add new strategies."
	}

	return status
}

func (m DeFiAgentTerminal) renderCyberpunkInput() string {
	// Cyberpunk input prompt with glitch effect
	inputContainer := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF00FF")).
		Padding(0, 1)

	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FFFF")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#1A1A1A")).
		Bold(true)

	prompt := promptStyle.Render("┃ AEGIS AI >")
	input := inputStyle.Render(m.Model.View())

	return inputContainer.Render(prompt + " " + input)
}

// autoCompleteCommand provides command autocompletion
func (m *DeFiAgentTerminal) autoCompleteCommand() {
	currentInput := m.Value()
	if currentInput == "" {
		return
	}

	availableCommands := []string{
		"deploy arbitrage",
		"check balance",
		"market status",
		"agent status",
		"start agent",
		"stop agent",
		"scan arbitrage",
		"yield rates",
		"gas status",
		"create strategy",
		"start strategies",
		"stop strategies",
		"strategy status",
		"help",
	}

	// Find matching commands
	var matches []string
	for _, cmd := range availableCommands {
		if strings.HasPrefix(cmd, currentInput) {
			matches = append(matches, cmd)
		}
	}

	// If only one match, complete it
	if len(matches) == 1 {
		m.SetValue(matches[0])
	} else if len(matches) > 1 {
		// Show available completions
		completionMsg := fmt.Sprintf("Available completions: %s", strings.Join(matches, ", "))
		m.messages = append(m.messages, completionMsg)
	}
}

func (m DeFiAgentTerminal) renderInput() string {
	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	return inputStyle.Render("🤖 Aegis AI > " + m.Model.View())
}

// updateMarketData fetches real-time market data from CoinGecko
func (m *DeFiAgentTerminal) updateMarketData() {
	// Fetch real prices from CoinGecko
	symbols := []string{"ETH/USD", "BTC/USD", "USDC/USD", "USDT/USD"}

	for _, symbol := range symbols {
		priceData, err := m.priceClient.GetPriceWithFallback(symbol)
		if err != nil {
			log.Printf("Failed to fetch price for %s: %v", symbol, err)
			continue
		}

		// Update market data with real prices
		switch symbol {
		case "ETH/USD":
			m.marketData.ETHPrice = priceData.Price
		case "BTC/USD":
			m.marketData.BTCPrice = priceData.Price
		case "USDC/USD":
			m.marketData.USDCPrice = priceData.Price
		case "USDT/USD":
			m.marketData.USDTPrice = priceData.Price
		}
	}

	// Update wallet total value based on real prices
	m.updateWalletValue()

	// Simulate realistic gas price changes
	gasChange := rand.Intn(8) - 4 // ±4 Gwei
	m.marketData.GasPrice += gasChange
	if m.marketData.GasPrice < 15 {
		m.marketData.GasPrice = 15
	} else if m.marketData.GasPrice > 80 {
		m.marketData.GasPrice = 80
	}

	// Update arbitrage opportunities
	m.updateArbitrageOpportunities()
}

// updateWalletValue recalculates wallet total value based on current market prices
func (m *DeFiAgentTerminal) updateWalletValue() {
	ethValue := m.walletData.ETHBalance * m.marketData.ETHPrice
	m.walletData.TotalValue = ethValue + m.walletData.USDCBalance
}

// simulateRealisticPrices generates realistic cryptocurrency prices
func (m *DeFiAgentTerminal) simulateRealisticPrices() {
	// Realistic price ranges
	ethChange := (rand.Float64() - 0.5) * 50  // ±$25
	btcChange := (rand.Float64() - 0.5) * 200 // ±$100

	m.marketData.ETHPrice += ethChange
	m.marketData.BTCPrice += btcChange

	// Keep prices in realistic ranges
	if m.marketData.ETHPrice < 3200 {
		m.marketData.ETHPrice = 3200 + rand.Float64()*200
	} else if m.marketData.ETHPrice > 4200 {
		m.marketData.ETHPrice = 4200 - rand.Float64()*200
	}

	if m.marketData.BTCPrice < 60000 {
		m.marketData.BTCPrice = 60000 + rand.Float64()*5000
	} else if m.marketData.BTCPrice > 75000 {
		m.marketData.BTCPrice = 75000 - rand.Float64()*5000
	}

	// Stablecoins should be very close to $1
	m.marketData.USDCPrice = 1.00
	m.marketData.USDTPrice = 0.9995 + rand.Float64()*0.001
}

// updateArbitrageOpportunities simulates finding new arbitrage opportunities
func (m *DeFiAgentTerminal) updateArbitrageOpportunities() {
	// Clear existing opportunities
	m.marketData.ArbitrageOps = []ArbitrageOpportunity{}

	// Simulate finding new opportunities
	if rand.Float64() > 0.3 { // 70% chance of finding an opportunity
		profitPct := 0.1 + rand.Float64()*0.5 // 0.1% to 0.6% profit
		m.marketData.ArbitrageOps = append(m.marketData.ArbitrageOps, ArbitrageOpportunity{
			Pair:      "ETH/USDC",
			ExchangeA: "Uniswap",
			ExchangeB: "PancakeSwap",
			PriceA:    m.marketData.ETHPrice * 0.998,
			PriceB:    m.marketData.ETHPrice * 1.002,
			Profit:    124.50,
			ProfitPct: profitPct,
		})
	}
}

func (m DeFiAgentTerminal) processCommand(command string) string {
	switch command {
	case "", "help":
		return `Available commands:
• deploy arbitrage - Start arbitrage detection
• check balance - Show wallet balances  
• market status - Current market conditions
• agent status - Agent performance
• start/stop agent <name> - Control agents
• scan arbitrage - Find opportunities
• yield rates - Yield farming rates
• gas status - Network conditions
• create strategy - New trading strategy
• start strategies - Start all strategies
• stop strategies - Stop all strategies
• strategy status - Show strategy status

Navigation:
• [1-6] Switch views (Dashboard, Agents, Market, Wallet, Strategies, Trading)
• [F] Toggle full screen mode
• [ESC] Return to dashboard
• [T] Quick trade (in trading view)`

	case "deploy arbitrage":
		return "🟢 Arbitrage agents deployed. Scanning for ETH/USDC and AVAX/USDT opportunities..."

	case "check balance":
		return fmt.Sprintf("💰 Total: $%.2f | ETH: %.2f ($%.2f) | USDC: %.2f",
			m.walletData.TotalValue, m.walletData.ETHBalance,
			m.walletData.ETHBalance*m.marketData.ETHPrice, m.walletData.USDCBalance)

	case "market status":
		return fmt.Sprintf("📊 ETH: $%.2f | BTC: $%.2f | %d arbitrage opportunities detected",
			m.marketData.ETHPrice, m.marketData.BTCPrice, len(m.marketData.ArbitrageOps))

	case "agent status":
		return "🤖 4 agents active | Efficiency: 98.7% | 47 trades executed | $245 profit"

	case "scan arbitrage":
		m.updateArbitrageOpportunities()
		if len(m.marketData.ArbitrageOps) > 0 {
			return fmt.Sprintf("🔍 Scanning... Found %d opportunities with up to %.2f%% profit",
				len(m.marketData.ArbitrageOps), m.marketData.ArbitrageOps[0].ProfitPct)
		}
		return "🔍 Scanning... No arbitrage opportunities found"

	case "yield rates":
		return fmt.Sprintf("🏦 Aave USDC: %.1f%% | Compound ETH: %.1f%% | Uniswap V3: %.1f%%",
			m.marketData.YieldRates["Aave USDC"],
			m.marketData.YieldRates["Compound ETH"],
			m.marketData.YieldRates["Uniswap V3"])

	case "gas status":
		return fmt.Sprintf("⛽ Current: %d Gwei | Recommended: Wait for <25 Gwei", m.marketData.GasPrice)

	case "create strategy":
		return "⚙️ Strategy creation interface activated. Use 'modify strategy' to configure parameters."

	case "start strategies":
		// Start the strategy engine
		ctx := context.Background()
		if err := m.strategyEngine.Start(ctx); err != nil {
			return fmt.Sprintf("❌ Failed to start strategies: %v", err)
		}
		return "🟢 Strategy engine started. Monitoring market conditions..."

	case "stop strategies":
		m.strategyEngine.Stop()
		return "🟡 Strategy engine stopped. All strategies paused."

	case "strategy status":
		return m.getStrategyStatus()

	default:
		return fmt.Sprintf("❌ Unknown command: %s. Type 'help' for available commands.", command)
	}
}

func main() {
	// Check if running in non-interactive mode
	if len(os.Args) > 1 && os.Args[1] == "--non-interactive" {
		runNonInteractiveMode()
		return
	}

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n\n🤖 Aegis DeFi AI Agent shutting down gracefully...")
		os.Exit(0)
	}()

	// Initialize and run the terminal
	p := tea.NewProgram(NewDeFiAgentTerminal(), tea.WithAltScreen())

	fmt.Println("🚀 Starting Aegis DeFi AI Agent Terminal...")
	fmt.Println("💡 Use number keys 1-6 to navigate between views")
	fmt.Println("💡 Type commands to interact with the DeFi AI Agent")
	fmt.Println("💡 Press 'F' for full screen mode, 'T' for quick trade")
	fmt.Println("💡 Press 'q' or Ctrl+C to exit")

	time.Sleep(2 * time.Second)

	if _, err := p.Run(); err != nil {
		log.Fatal("Error running terminal:", err)
	}
}

// runNonInteractiveMode runs the agent in non-interactive mode
func runNonInteractiveMode() {
	terminal := NewDeFiAgentTerminal()

	if len(os.Args) > 2 {
		switch os.Args[2] {
		case "--status":
			fmt.Println("📊 System Status:")
			fmt.Println(terminal.renderDashboard())
			return
		case "--market":
			fmt.Println("📈 Market Data:")
			fmt.Println(terminal.renderMarket())
			return
		case "--strategies":
			fmt.Println("🤖 Strategy Status:")
			fmt.Println(terminal.getStrategyStatus())
			return
		case "--agents":
			fmt.Println("👥 Agent Status:")
			fmt.Println(terminal.renderAgents())
			return
		case "--wallet":
			fmt.Println("💰 Wallet Status:")
			fmt.Println(terminal.renderWallet())
			return
		case "--trading":
			fmt.Println("📈 Trading Interface:")
			fmt.Println(terminal.renderTradingView())
			return
		}
	}

	// Default: show all information
	fmt.Println("🤖 Aegis DeFi AI Agent - Non-Interactive Mode")
	fmt.Println("=============================================")

	fmt.Println("\n📊 System Status:")
	fmt.Println(terminal.renderDashboard())

	fmt.Println("\n📈 Market Data:")
	fmt.Println(terminal.renderMarket())

	fmt.Println("\n🤖 Strategy Status:")
	fmt.Println(terminal.getStrategyStatus())

	fmt.Println("\n💡 Available Commands:")
	fmt.Println("  ./bin/aegis-terminal --non-interactive")
	fmt.Println("  ./bin/aegis-terminal --status")
	fmt.Println("  ./bin/aegis-terminal --market")
	fmt.Println("  ./bin/aegis-terminal --strategies")
	fmt.Println("  ./bin/aegis-terminal --agents")
	fmt.Println("  ./bin/aegis-terminal --wallet")
}

// renderTradingView renders the trading interface with price charts
func (m DeFiAgentTerminal) renderTradingView() string {
	// Update price history with current price
	currentPrice := m.marketData.ETHPrice
	if len(m.tradingData.PriceHistory) > 0 {
		lastPrice := m.tradingData.PriceHistory[len(m.tradingData.PriceHistory)-1].Price
		// Add new price point if significant change or time elapsed
		if math.Abs(currentPrice-lastPrice) > 1.0 ||
			time.Since(m.tradingData.PriceHistory[len(m.tradingData.PriceHistory)-1].Timestamp) > time.Minute {
			m.tradingData.PriceHistory = append(m.tradingData.PriceHistory, PricePoint{
				Timestamp: time.Now(),
				Price:     currentPrice,
				Volume:    rand.Float64() * 1000,
			})
			// Keep only last 50 points
			if len(m.tradingData.PriceHistory) > 50 {
				m.tradingData.PriceHistory = m.tradingData.PriceHistory[1:]
			}
		}
	}

	// Calculate chart statistics
	var minPrice, maxPrice float64
	if len(m.tradingData.PriceHistory) > 0 {
		minPrice = m.tradingData.PriceHistory[0].Price
		maxPrice = m.tradingData.PriceHistory[0].Price
		for _, point := range m.tradingData.PriceHistory {
			if point.Price < minPrice {
				minPrice = point.Price
			}
			if point.Price > maxPrice {
				maxPrice = point.Price
			}
		}
	}

	// Render price chart
	chart := m.renderPriceChart(minPrice, maxPrice)

	return fmt.Sprintf(`
┌─────────────────────────────────────────────────────────────────────────┐
│                           TRADING INTERFACE                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  📊 ACTIVE PAIR: %s                                                    │
│  💰 CURRENT PRICE: $%.2f                                               │
│  📈 24H CHANGE: %.2f%%                                                 │
│                                                                         │
│  %s
│                                                                         │
│  🎯 QUICK ACTIONS:                                                      │
│     [B] Buy ETH    [S] Sell ETH    [T] Quick Trade                      │
│     [1] ETH/USDC   [2] BTC/USD     [3] SOL/USD                          │
│                                                                         │
│  📋 RECENT TRADES:                                                      │
%s│                                                                         │
│  💡 Commands:                                                           │
│     • 'buy <amount>' - Buy cryptocurrency                              │
│     • 'sell <amount>' - Sell cryptocurrency                            │
│     • 'switch <pair>' - Switch trading pair                            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`,
		m.tradingData.ActivePair,
		currentPrice,
		(currentPrice-3800.0)/3800.0*100,
		chart,
		m.renderRecentTrades())
}

// renderPriceChart creates an ASCII price chart
func (m DeFiAgentTerminal) renderPriceChart(minPrice, maxPrice float64) string {
	if len(m.tradingData.PriceHistory) < 2 {
		return "     No price data available"
	}

	chartHeight := 10
	chartWidth := len(m.tradingData.PriceHistory)
	if chartWidth > 40 {
		chartWidth = 40
	}

	// Create chart grid
	chart := make([][]rune, chartHeight)
	for i := range chart {
		chart[i] = make([]rune, chartWidth)
		for j := range chart[i] {
			chart[i][j] = ' '
		}
	}

	// Plot price points
	priceRange := maxPrice - minPrice
	if priceRange == 0 {
		priceRange = 1
	}

	for i, point := range m.tradingData.PriceHistory[len(m.tradingData.PriceHistory)-chartWidth:] {
		if i >= chartWidth {
			break
		}
		y := int(float64(chartHeight-1) * (1 - (point.Price-minPrice)/priceRange))
		if y >= 0 && y < chartHeight {
			chart[y][i] = '●'
		}
	}

	// Convert chart to string
	var chartStr strings.Builder
	chartStr.WriteString("     ┌" + strings.Repeat("─", chartWidth) + "┐\n")
	for i := 0; i < chartHeight; i++ {
		chartStr.WriteString("     │")
		for j := 0; j < chartWidth; j++ {
			chartStr.WriteRune(chart[i][j])
		}
		chartStr.WriteString("│\n")
	}
	chartStr.WriteString("     └" + strings.Repeat("─", chartWidth) + "┘")
	chartStr.WriteString(fmt.Sprintf("\n     Min: $%.0f  Max: $%.0f", minPrice, maxPrice))

	return chartStr.String()
}

// renderRecentTrades shows recent trading activity
func (m DeFiAgentTerminal) renderRecentTrades() string {
	if len(m.tradingData.TradeHistory) == 0 {
		return "     No recent trades"
	}

	var trades strings.Builder
	// Show last 3 trades
	start := len(m.tradingData.TradeHistory) - 3
	if start < 0 {
		start = 0
	}

	for i := start; i < len(m.tradingData.TradeHistory); i++ {
		trade := m.tradingData.TradeHistory[i]
		profitStr := ""
		if trade.Profit > 0 {
			profitStr = fmt.Sprintf("(+$%.2f)", trade.Profit)
		} else if trade.Profit < 0 {
			profitStr = fmt.Sprintf("(-$%.2f)", math.Abs(trade.Profit))
		}
		trades.WriteString(fmt.Sprintf("     • %s %.2f %s @ $%.2f %s\n",
			strings.ToUpper(trade.Type), trade.Amount, trade.Pair, trade.Price, profitStr))
	}

	return trades.String()
}

// executeQuickTrade executes a quick trade with default parameters
func (m *DeFiAgentTerminal) executeQuickTrade() {
	currentPrice := m.marketData.ETHPrice
	amount := 0.1
	cost := amount * currentPrice

	// Check if we have enough balance
	if m.tradingData.CurrentBalance >= cost {
		// Execute buy
		m.tradingData.CurrentBalance -= cost
		m.walletData.ETHBalance += amount

		// Record trade
		trade := Trade{
			ID:        fmt.Sprintf("QT-%d", time.Now().Unix()),
			Pair:      m.tradingData.ActivePair,
			Type:      "buy",
			Amount:    amount,
			Price:     currentPrice,
			Timestamp: time.Now(),
			Profit:    0.0,
		}
		m.tradingData.TradeHistory = append(m.tradingData.TradeHistory, trade)

		m.messages = append(m.messages, fmt.Sprintf("⚡ Quick trade executed: BUY %.2f %s @ $%.2f", amount, m.tradingData.ActivePair, currentPrice))
	} else {
		m.messages = append(m.messages, "❌ Insufficient balance for quick trade")
	}
}

// executeBuy executes a buy order
func (m *DeFiAgentTerminal) executeBuy(amount float64) {
	currentPrice := m.getCurrentPriceForPair()
	cost := amount * currentPrice

	// Check if we have enough balance
	if m.tradingData.CurrentBalance >= cost {
		// Execute buy
		m.tradingData.CurrentBalance -= cost
		m.updateAssetBalance(m.tradingData.ActivePair, amount, true)

		// Record trade
		trade := Trade{
			ID:        fmt.Sprintf("BUY-%d", time.Now().Unix()),
			Pair:      m.tradingData.ActivePair,
			Type:      "buy",
			Amount:    amount,
			Price:     currentPrice,
			Timestamp: time.Now(),
			Profit:    0.0,
		}
		m.tradingData.TradeHistory = append(m.tradingData.TradeHistory, trade)

		m.messages = append(m.messages, fmt.Sprintf("🟢 BUY order executed: %.2f %s @ $%.2f", amount, m.tradingData.ActivePair, currentPrice))
	} else {
		m.messages = append(m.messages, fmt.Sprintf("❌ Insufficient balance. Need $%.2f, have $%.2f", cost, m.tradingData.CurrentBalance))
	}
}

// executeSell executes a sell order
func (m *DeFiAgentTerminal) executeSell(amount float64) {
	currentPrice := m.getCurrentPriceForPair()
	revenue := amount * currentPrice

	// Check if we have enough asset
	if m.getAssetBalance(m.tradingData.ActivePair) >= amount {
		// Execute sell
		m.tradingData.CurrentBalance += revenue
		m.updateAssetBalance(m.tradingData.ActivePair, amount, false)

		// Calculate profit (simplified)
		profit := revenue - (amount * currentPrice * 0.99) // Assume 1% profit

		// Record trade
		trade := Trade{
			ID:        fmt.Sprintf("SELL-%d", time.Now().Unix()),
			Pair:      m.tradingData.ActivePair,
			Type:      "sell",
			Amount:    amount,
			Price:     currentPrice,
			Timestamp: time.Now(),
			Profit:    profit,
		}
		m.tradingData.TradeHistory = append(m.tradingData.TradeHistory, trade)

		m.messages = append(m.messages, fmt.Sprintf("🔴 SELL order executed: %.2f %s @ $%.2f (Profit: $%.2f)", amount, m.tradingData.ActivePair, currentPrice, profit))
	} else {
		m.messages = append(m.messages, fmt.Sprintf("❌ Insufficient %s balance. Need %.2f, have %.2f", m.tradingData.ActivePair, amount, m.getAssetBalance(m.tradingData.ActivePair)))
	}
}

// getCurrentPriceForPair returns the current price for the active trading pair
func (m *DeFiAgentTerminal) getCurrentPriceForPair() float64 {
	switch m.tradingData.ActivePair {
	case "ETH/USDC":
		return m.marketData.ETHPrice
	case "BTC/USD":
		return m.marketData.BTCPrice
	case "SOL/USD":
		// Simulate SOL price
		return 150.0 + rand.Float64()*50
	default:
		return m.marketData.ETHPrice
	}
}

// getAssetBalance returns the balance for the given trading pair
func (m *DeFiAgentTerminal) getAssetBalance(pair string) float64 {
	switch pair {
	case "ETH/USDC":
		return m.walletData.ETHBalance
	case "BTC/USD":
		// Simulate BTC balance
		return 0.05
	case "SOL/USD":
		// Simulate SOL balance
		return 5.0
	default:
		return m.walletData.ETHBalance
	}
}

// updateAssetBalance updates the balance for the given trading pair
func (m *DeFiAgentTerminal) updateAssetBalance(pair string, amount float64, isBuy bool) {
	if isBuy {
		// Buying: increase asset balance
		switch pair {
		case "ETH/USDC":
			m.walletData.ETHBalance += amount
			// BTC and SOL balances are simulated and not stored in wallet data
		}
	} else {
		// Selling: decrease asset balance
		switch pair {
		case "ETH/USDC":
			m.walletData.ETHBalance -= amount
			// BTC and SOL balances are simulated and not stored in wallet data
		}
	}
}
