package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/agent"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/market"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/wallet"
)

type ViewType int

const (
	WalletView ViewType = iota
	MarketView
	AgentView
	ChatView
	SettingsView
)

type NexusAIModel struct {
	// Core Application State
	currentView   ViewType
	walletManager *wallet.Manager
	marketData    *market.Data
	agentManager  *agent.Manager

	// TUI Components
	walletSelector list.Model
	marketViewer   viewport.Model
	chatInput      textinput.Model
	statusBar      StatusBar
	spinner        spinner.Model

	// Navigation
	cursor         int
	selectedWallet string

	// Commands
	pendingCmds []tea.Cmd

	// UI State
	width   int
	height  int
	loading bool
}

type StatusBar struct {
	message string
	style   lipgloss.Style
}

func NewNexusAIModel() NexusAIModel {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	// Initialize wallet selector
	walletList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	walletList.Title = "Wallets"
	walletList.SetShowStatusBar(false)
	walletList.SetFilteringEnabled(false)

	// Initialize market viewer
	marketViewer := viewport.New(0, 0)
	marketViewer.SetContent("Loading market data...")

	// Initialize chat input
	chatInput := textinput.New()
	chatInput.Placeholder = "Ask NexusAI about DeFi opportunities..."
	chatInput.Focus()
	chatInput.CharLimit = 256
	chatInput.Width = 50

	return NexusAIModel{
		currentView:    WalletView,
		walletSelector: walletList,
		marketViewer:   marketViewer,
		chatInput:      chatInput,
		spinner:        s,
		statusBar: StatusBar{
			message: "Ready",
			style:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
		},
		loading: false,
	}
}

func (m NexusAIModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.loadInitialData(),
	)
}

func (m NexusAIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeComponents()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.currentView = WalletView
		case "2":
			m.currentView = MarketView
		case "3":
			m.currentView = AgentView
		case "4":
			m.currentView = ChatView
		case "5":
			m.currentView = SettingsView
		case "tab":
			m.cycleView()
		case "r":
			if m.currentView == WalletView && m.walletManager != nil {
				m.walletManager.RefreshBalances()
				m.statusBar.message = "Balances refreshed"
				// Update wallet list
				var walletItems []list.Item
				for _, w := range m.walletManager.GetWallets() {
					walletItems = append(walletItems, WalletItem{
						Address:  w.Address,
						Chain:    w.Chain,
						Balance:  w.Balance,
						IsActive: w.IsActive,
					})
				}
				m.walletSelector.SetItems(walletItems)
			} else if m.currentView == MarketView && m.marketData != nil {
				m.marketData.UpdatePrices()
				m.statusBar.message = "Market data updated"
				m.marketViewer.SetContent(m.renderMarketData())
			}
		case "enter":
			if m.currentView == WalletView {
				if selected := m.walletSelector.SelectedItem(); selected != nil {
					if walletItem, ok := selected.(WalletItem); ok {
						m.walletManager.SelectWallet(walletItem.Address)
						m.statusBar.message = fmt.Sprintf("Selected wallet: %s", walletItem.Address[:16]+"...")
						// Update wallet list to reflect selection
						var walletItems []list.Item
						for _, w := range m.walletManager.GetWallets() {
							walletItems = append(walletItems, WalletItem{
								Address:  w.Address,
								Chain:    w.Chain,
								Balance:  w.Balance,
								IsActive: w.IsActive,
							})
						}
						m.walletSelector.SetItems(walletItems)
					}
				}
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case DataLoadedMsg:
		m.loading = false
		m.statusBar.message = "Data loaded successfully"
	}

	// Update components based on current view
	switch m.currentView {
	case WalletView:
		var cmd tea.Cmd
		m.walletSelector, cmd = m.walletSelector.Update(msg)
		cmds = append(cmds, cmd)
	case MarketView:
		var cmd tea.Cmd
		m.marketViewer, cmd = m.marketViewer.Update(msg)
		cmds = append(cmds, cmd)
	case ChatView:
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m NexusAIModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	var content string

	switch m.currentView {
	case WalletView:
		content = m.renderWalletView()
	case MarketView:
		content = m.renderMarketView()
	case AgentView:
		content = m.renderAgentView()
	case ChatView:
		content = m.renderChatView()
	case SettingsView:
		content = m.renderSettingsView()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		content,
		m.renderStatusBar(),
	)
}

func (m *NexusAIModel) resizeComponents() {
	contentHeight := m.height - 4 // Header + status bar
	contentWidth := m.width - 2   // Padding

	// Resize wallet selector
	m.walletSelector.SetSize(contentWidth, contentHeight)

	// Resize market viewer
	m.marketViewer.Width = contentWidth
	m.marketViewer.Height = contentHeight

	// Resize chat input
	m.chatInput.Width = contentWidth - 4
}

func (m *NexusAIModel) cycleView() {
	m.currentView = (m.currentView + 1) % 5
}

func (m *NexusAIModel) loadInitialData() tea.Cmd {
	m.loading = true
	m.statusBar.message = "Loading initial data..."

	// Initialize managers
	m.walletManager = wallet.NewManager()
	m.marketData = market.NewData()
	m.agentManager = agent.NewManager()

	// Populate wallet list
	var walletItems []list.Item
	for _, w := range m.walletManager.GetWallets() {
		walletItems = append(walletItems, WalletItem{
			Address:  w.Address,
			Chain:    w.Chain,
			Balance:  w.Balance,
			IsActive: w.IsActive,
		})
	}
	m.walletSelector.SetItems(walletItems)

	// Set market data
	m.marketViewer.SetContent(m.renderMarketData())

	return func() tea.Msg {
		time.Sleep(1 * time.Second) // Simulate loading
		return DataLoadedMsg{}
	}
}

// Custom message types
type DataLoadedMsg struct{}

// Wallet item for list component
type WalletItem struct {
	Address  string
	Chain    string
	Balance  string
	IsActive bool
}

func (w WalletItem) FilterValue() string { return w.Address }
func (w WalletItem) Title() string {
	if w.IsActive {
		return fmt.Sprintf("🟢 %s (%s)", w.Address[:16]+"...", w.Chain)
	}
	return fmt.Sprintf("⚪ %s (%s)", w.Address[:16]+"...", w.Chain)
}
func (w WalletItem) Description() string {
	return fmt.Sprintf("Balance: %s", w.Balance)
}

// View rendering methods
func (m NexusAIModel) renderHeader() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(0, 1)

	views := []string{
		fmt.Sprintf("1. %s", m.viewIndicator("Wallets", WalletView)),
		fmt.Sprintf("2. %s", m.viewIndicator("Market", MarketView)),
		fmt.Sprintf("3. %s", m.viewIndicator("Agents", AgentView)),
		fmt.Sprintf("4. %s", m.viewIndicator("Chat", ChatView)),
		fmt.Sprintf("5. %s", m.viewIndicator("Settings", SettingsView)),
	}

	return headerStyle.Render("NexusAI - " + lipgloss.JoinHorizontal(lipgloss.Left, views...))
}

func (m NexusAIModel) viewIndicator(name string, view ViewType) string {
	if m.currentView == view {
		return fmt.Sprintf("[%s]", name)
	}
	return name
}

func (m NexusAIModel) renderWalletView() string {
	if m.loading {
		return fmt.Sprintf("\n%s Loading wallets...", m.spinner.View())
	}

	var content string

	// Show wallet list
	content += fmt.Sprintf("\n%s", m.walletSelector.View())

	// Show detailed wallet info if a wallet is selected
	if m.walletManager != nil && m.walletManager.GetCurrentWallet() != nil {
		currentWallet := m.walletManager.GetCurrentWallet()
		content += "\n\n📋 Selected Wallet Details:\n"
		content += fmt.Sprintf("  Address: %s\n", currentWallet.Address)
		content += fmt.Sprintf("  Chain: %s\n", currentWallet.Chain)
		content += fmt.Sprintf("  Total Balance: %s\n", currentWallet.Balance)

		content += "\n  Token Balances:\n"
		totalUSD := 0.0
		for _, token := range currentWallet.Tokens {
			content += fmt.Sprintf("    %-6s: %-12s ($%s)\n", token.Symbol, token.Balance, token.ValueUSD)
			// Calculate total USD value
			var value float64
			fmt.Sscanf(token.ValueUSD, "%f", &value)
			totalUSD += value
		}
		content += fmt.Sprintf("\n  Total Value: $%.2f\n", totalUSD)
	}

	// Add help text
	content += "\n\n💡 Help: Press 'r' to refresh balances, 'enter' to select wallet"

	return content
}

func (m NexusAIModel) renderMarketView() string {
	if m.loading {
		return fmt.Sprintf("\n%s Loading market data...", m.spinner.View())
	}
	return fmt.Sprintf("\n%s", m.marketViewer.View())
}

func (m NexusAIModel) renderAgentView() string {
	return "\nAgent Management View\n\nComing soon: AI agent coordination and automation"
}

func (m NexusAIModel) renderChatView() string {
	return fmt.Sprintf(
		"\nChat with NexusAI\n\n%s\n\nType your question about DeFi opportunities...",
		m.chatInput.View(),
	)
}

func (m NexusAIModel) renderSettingsView() string {
	return "\nSettings View\n\nConfigure your NexusAI preferences and integrations"
}

func (m NexusAIModel) renderStatusBar() string {
	status := m.statusBar.message
	if m.loading {
		status = fmt.Sprintf("%s %s", m.spinner.View(), status)
	}
	return m.statusBar.style.Render(status)
}

func (m NexusAIModel) renderMarketData() string {
	if m.marketData == nil {
		return "Loading market data..."
	}

	var content string
	content += fmt.Sprintf("\n📊 Market Prices (Last Updated: %s)\n\n", m.marketData.GetLastUpdate().Format("15:04:05"))

	for symbol, price := range m.marketData.GetAllPrices() {
		changeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
		if price.Change24h < 0 {
			changeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("red"))
		}

		changeStr := fmt.Sprintf("%.2f%%", price.Change24h)
		if price.Change24h > 0 {
			changeStr = "+" + changeStr
		}

		content += fmt.Sprintf("%-6s $%-10.2f %-12s Vol: $%.1fB\n",
			symbol,
			price.Price,
			changeStyle.Render(changeStr),
			price.Volume/1e9)
	}

	content += "\n🏦 DeFi Protocols\n\n"
	for _, protocol := range m.marketData.GetProtocols() {
		content += fmt.Sprintf("%-15s TVL: $%.1fB APY: %.1f%% %s\n",
			protocol.Name,
			protocol.TVL/1e9,
			protocol.APY,
			protocol.Category)
	}

	content += "\n\n💡 Help: Press 'r' to refresh market data"

	return content
}
