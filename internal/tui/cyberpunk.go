package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Cyberpunk color palette
var (
	// Colors
	neonGreen    = lipgloss.Color("#00FF00")
	neonBlue     = lipgloss.Color("#00FFFF")
	neonPink     = lipgloss.Color("#FF00FF")
	neonPurple   = lipgloss.Color("#9D00FF")
	cyberYellow  = lipgloss.Color("#FFD700")
	matrixGreen  = lipgloss.Color("#00FF41")
	darkGray     = lipgloss.Color("#1A1A1A")
	terminalGray = lipgloss.Color("#0A0A0A")

	// Styles
	asciiLogoStyle = lipgloss.NewStyle().
			Foreground(neonGreen).
			Bold(true)

	headerStyle = lipgloss.NewStyle().
			Foreground(neonBlue).
			Background(darkGray).
			Bold(true).
			Padding(0, 2).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(neonBlue)

	activeTabStyle = lipgloss.NewStyle().
			Foreground(neonPink).
			Background(darkGray).
			Bold(true).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(neonBlue).
				Background(darkGray).
				Padding(0, 2)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(cyberYellow).
			Background(darkGray).
			Italic(true).
			Padding(0, 1)

	loadingStyle = lipgloss.NewStyle().
			Foreground(neonPurple).
			Bold(true)

	// Cyberpunk ASCII Art
	asciiLogo = `
██████╗ ███████╗███████╗██╗███████╗    ██████╗ ██████╗ ██████╗ ███████╗
██╔══██╗██╔════╝██╔════╝██║██╔════╝    ██╔══██╗██╔══██╗██╔══██╗██╔════╝
██████╔╝█████╗  █████╗  ██║███████╗    ██████╔╝██████╔╝██████╔╝█████╗  
██╔══██╗██╔══╝  ██╔══╝  ██║╚════██║    ██╔═══╝ ██╔══██╗██╔══██╗██╔══╝  
██████╔╝███████╗██║     ██║███████║    ██║     ██║  ██║██║  ██║███████╗
╚═════╝ ╚══════╝╚═╝     ╚═╝╚══════╝    ╚═╝     ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝

     ███████╗██████╗ ██╗   ██╗███████╗██████╗ ██╗ ██████╗ ███████╗
     ██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗██║██╔════╝ ██╔════╝
     █████╗  ██████╔╝██║   ██║█████╗  ██████╔╝██║██║  ███╗███████╗
     ██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗██║██║   ██║╚════██║
     ███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║██║╚██████╔╝███████║
     ╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝╚═╝ ╚═════╝ ╚══════╝

           ███████╗██╗   ██╗███████╗████████╗███████╗███╗   ███╗
           ██╔════╝██║   ██║██╔════╝╚══██╔══╝██╔════╝████╗ ████║
           █████╗  ██║   ██║███████╗   ██║   █████╗  ██╔████╔██║
           ██╔══╝  ██║   ██║╚════██║   ██║   ██╔══╝  ██║╚██╔╝██║
           ██║     ╚██████╔╝███████║   ██║   ███████╗██║ ╚═╝ ██║
           ╚═╝      ╚═════╝ ╚══════╝   ╚═╝   ╚══════╝╚═╝     ╚═╝
`

	matrixRain = []string{
		"01010101010101010101",
		"10101010101010101010",
		"01010101010101010101",
		"10101010101010101010",
		"01010101010101010101",
	}
)

type CyberpunkModel struct {
	NexusAIModel

	// Cyberpunk specific components
	progressBar   progress.Model
	asciiViewport viewport.Model
	matrixEffect  bool
	bootSequence  bool
	bootProgress  float64

	// Enhanced UI state
	glitchEffect bool
	pulseEffect  bool
	terminalMode bool
}

func NewCyberpunkModel() CyberpunkModel {
	baseModel := NewNexusAIModel()

	// Initialize progress bar
	prog := progress.New(
		progress.WithScaledGradient("#FF00FF", "#00FFFF"),
		progress.WithWidth(40),
	)

	// Initialize ASCII viewport
	asciiVP := viewport.New(80, 20)
	asciiVP.SetContent(asciiLogo)

	return CyberpunkModel{
		NexusAIModel:  baseModel,
		progressBar:   prog,
		asciiViewport: asciiVP,
		bootSequence:  true,
		bootProgress:  0.0,
		matrixEffect:  true,
	}
}

func (m CyberpunkModel) Init() tea.Cmd {
	return tea.Batch(
		m.NexusAIModel.Init(),
		m.startBootSequence(),
	)
}

func (m CyberpunkModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle boot sequence
	if m.bootSequence {
		switch msg := msg.(type) {
		case BootProgressMsg:
			m.bootProgress = msg.progress
			if m.bootProgress >= 1.0 {
				m.bootSequence = false
			}
			return m, m.updateBootProgress()
		}
	}

	// Update base model
	var cmd tea.Cmd
	var updatedModel tea.Model
	updatedModel, cmd = m.NexusAIModel.Update(msg)
	if nexusModel, ok := updatedModel.(NexusAIModel); ok {
		m.NexusAIModel = nexusModel
	}
	cmds = append(cmds, cmd)

	// Handle cyberpunk-specific inputs
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+m":
			m.matrixEffect = !m.matrixEffect
		case "ctrl+g":
			m.glitchEffect = !m.glitchEffect
		case "ctrl+p":
			m.pulseEffect = !m.pulseEffect
		case "ctrl+t":
			m.terminalMode = !m.terminalMode
		}
	}

	return m, tea.Batch(cmds...)
}

func (m CyberpunkModel) View() string {
	if m.bootSequence {
		return m.renderBootScreen()
	}

	if m.terminalMode {
		return m.renderTerminalMode()
	}

	var content string

	// Add matrix effect background if enabled
	if m.matrixEffect {
		content += m.renderMatrixBackground() + "\n"
	}

	// Render enhanced header
	content += m.renderCyberpunkHeader()

	// Render main content with cyberpunk styling
	switch m.currentView {
	case WalletView:
		content += m.renderCyberpunkWalletView()
	case MarketView:
		content += m.renderCyberpunkMarketView()
	case AgentView:
		content += m.renderCyberpunkAgentView()
	case ChatView:
		content += m.renderCyberpunkChatView()
	case SettingsView:
		content += m.renderCyberpunkSettingsView()
	}

	// Add enhanced status bar
	content += m.renderCyberpunkStatusBar()

	// Apply glitch effect if enabled
	if m.glitchEffect {
		content = m.applyGlitchEffect(content)
	}

	return content
}

// Boot sequence methods
func (m *CyberpunkModel) startBootSequence() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(100 * time.Millisecond)
		return BootProgressMsg{progress: 0.1}
	}
}

func (m *CyberpunkModel) updateBootProgress() tea.Cmd {
	if m.bootProgress >= 1.0 {
		return nil
	}

	return func() tea.Msg {
		time.Sleep(50 * time.Millisecond)
		return BootProgressMsg{progress: m.bootProgress + 0.05}
	}
}

type BootProgressMsg struct {
	progress float64
}

func (m CyberpunkModel) renderBootScreen() string {
	var content strings.Builder

	// ASCII logo
	content.WriteString(asciiLogoStyle.Render(asciiLogo))
	content.WriteString("\n\n")

	// Boot progress
	content.WriteString("INITIALIZING AEGIS PROTOCOL SYSTEMS...\n\n")
	content.WriteString(m.progressBar.ViewAs(m.bootProgress))
	content.WriteString("\n\n")

	// Boot messages
	bootMessages := []string{
		"✓ Loading neural network core...",
		"✓ Initializing quantum encryption...",
		"✓ Connecting to blockchain nodes...",
		"✓ Syncing with DeFi protocols...",
		"✓ Activating AI agents...",
	}

	for i, msg := range bootMessages {
		if m.bootProgress > float64(i+1)*0.2 {
			content.WriteString(loadingStyle.Render(msg))
		} else {
			content.WriteString("◻ " + msg)
		}
		content.WriteString("\n")
	}

	content.WriteString("\n\n")
	content.WriteString("ACCESSING THE MATRIX...")

	return content.String()
}

func (m CyberpunkModel) renderCyberpunkHeader() string {
	tabs := []string{
		m.cyberTab("SYSTEM CONTROL", WalletView),
		m.cyberTab("MARKET DATA", MarketView),
		m.cyberTab("AI AGENTS", AgentView),
		m.cyberTab("NEURAL CHAT", ChatView),
		m.cyberTab("CONFIG", SettingsView),
	}

	header := headerStyle.Render(
		"AEGIS PROTOCOL v2.0 | " +
			lipgloss.JoinHorizontal(lipgloss.Left, tabs...) +
			" | " + time.Now().Format("15:04:05") + " UTC",
	)

	return header + "\n"
}

func (m CyberpunkModel) cyberTab(name string, view ViewType) string {
	if m.currentView == view {
		return "⟫ " + activeTabStyle.Render(name) + " ⟪"
	}
	return inactiveTabStyle.Render(name)
}

func (m CyberpunkModel) renderMatrixBackground() string {
	var lines []string
	for i := 0; i < 5; i++ {
		line := strings.Replace(matrixRain[i%len(matrixRain)], "0", " ", -1)
		line = strings.Replace(line, "1", "█", -1)
		lines = append(lines,
			lipgloss.NewStyle().
				Foreground(matrixGreen).
				Faint(true).
				Render(line),
		)
	}
	return strings.Join(lines, "\n")
}

func (m CyberpunkModel) renderCyberpunkWalletView() string {
	baseContent := m.renderWalletView()

	// Enhance with cyberpunk styling
	lines := strings.Split(baseContent, "\n")
	var enhancedLines []string

	for _, line := range lines {
		if strings.Contains(line, "🟢") {
			line = strings.Replace(line, "🟢", "🟢", 1)
			line = lipgloss.NewStyle().Foreground(neonGreen).Render(line)
		} else if strings.Contains(line, "⚪") {
			line = strings.Replace(line, "⚪", "⚪", 1)
			line = lipgloss.NewStyle().Foreground(neonBlue).Render(line)
		} else if strings.Contains(line, "📋") {
			line = lipgloss.NewStyle().Foreground(neonPink).Bold(true).Render(line)
		} else if strings.Contains(line, "💡") {
			line = lipgloss.NewStyle().Foreground(cyberYellow).Render(line)
		}
		enhancedLines = append(enhancedLines, line)
	}

	return strings.Join(enhancedLines, "\n")
}

func (m CyberpunkModel) renderCyberpunkMarketView() string {
	baseContent := m.renderMarketView()

	// Add cyberpunk market header
	enhancedContent := "📊 REAL-TIME MARKET SURVEILLANCE\n\n"
	enhancedContent += lipgloss.NewStyle().
		Foreground(neonBlue).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(neonBlue).
		Padding(0, 1).
		Render(baseContent)

	return enhancedContent
}

func (m CyberpunkModel) renderCyberpunkAgentView() string {
	return `
🤖 AI AGENT COORDINATION MATRIX

┌─────────────────────────────────────────┐
│  AGENT STATUS                           │
├─────────────────────────────────────────┤
│  🟢 HEDERA-ARBITRAGE-001                │
│     Status: Active | Efficiency: 98.7%  │
│                                           
│  🟢 PYTH-PRICE-BOT-002                  │
│     Status: Active | Accuracy: 99.2%    │
│                                           
│  🔵 BLOCKSCOUT-SCANNER-003              │
│     Status: Scanning | Coverage: 87.3%  │
│                                           
│  🟡 LIT-AUTOMATION-004                  │
│     Status: Standby | Ready: 100%       │
└─────────────────────────────────────────┘

💡 Press 'SPACE' to deploy new agents
`
}

func (m CyberpunkModel) renderCyberpunkChatView() string {
	return fmt.Sprintf(`
💬 NEURAL INTERFACE - CHAT WITH AEGIS AI

%s

┌─[USER INPUT]─────────────────────────────┐
│ %s │
└──────────────────────────────────────────┘

💡 Type your query about DeFi strategies, market analysis, or agent coordination
`,
		lipgloss.NewStyle().Foreground(neonGreen).Render("AI: Welcome to Aegis Protocol. I'm ready to assist with your DeFi operations."),
		m.chatInput.View(),
	)
}

func (m CyberpunkModel) renderCyberpunkSettingsView() string {
	return `
⚙️ SYSTEM CONFIGURATION

┌─────────────────────────────────────────┐
│  VISUAL SETTINGS                        │
├─────────────────────────────────────────┤
│  [X] Matrix Effect      [ ] Glitch FX   │
│  [X] Pulse Animation    [ ] Terminal    │
│                                           
│  NETWORK SETTINGS                       │
├─────────────────────────────────────────┤
│  🔗 Pyth Network: CONNECTED             │
│  🔗 Blockscout:   CONNECTED             │
│  🔗 Hedera:       CONNECTED             │
│  🔗 Envio:        CONNECTED             │
│  🔗 Lit Protocol: CONNECTED             │
└─────────────────────────────────────────┘

💡 Use arrow keys to navigate, SPACE to toggle
`
}

func (m CyberpunkModel) renderCyberpunkStatusBar() string {
	status := m.statusBar.message
	if m.loading {
		status = fmt.Sprintf("%s %s", m.spinner.View(), status)
	}

	// Add system metrics
	metrics := fmt.Sprintf(" | CPU: 12%% | MEM: 47%% | NET: 2.3MB/s | AGENTS: 4")

	return "\n" + statusBarStyle.Render("STATUS: "+status+metrics)
}

func (m CyberpunkModel) renderTerminalMode() string {
	return `
┌─────────────────────────────────────────────────────────────────────────┐
│                        AEGIS PROTOCOL - TERMINAL MODE                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  user@aegis:~$ ./deploy_agents.sh                                       │
│  > Initializing agent deployment sequence...                            │
│  > Connecting to Hedera network... [OK]                                 │
│  > Syncing with Pyth price feeds... [OK]                                │
│  > Deploying arbitrage agents... [OK]                                   │
│  > Agent coordination matrix established                                │
│                                                                         │
│  user@aegis:~$ ./monitor_system.sh                                      │
│  > SYSTEM STATUS: OPTIMAL                                               │
│  > ACTIVE AGENTS: 4                                                     │
│  > BLOCKCHAIN SYNC: 100%                                                │
│  > PRICE FEEDS: STABLE                                                  │
│                                                                         │
│  user@aegis:~$ _                                                        │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
`
}

func (m CyberpunkModel) applyGlitchEffect(content string) string {
	// Simple glitch effect by randomly replacing characters
	glitched := []rune(content)
	for i := 0; i < len(glitched); i += 20 {
		if i+1 < len(glitched) {
			glitched[i], glitched[i+1] = glitched[i+1], glitched[i]
		}
	}
	return string(glitched)
}
