package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:     "nexusai",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Short:   "NexusAI - Intelligent Cross-Chain DeFi Terminal",
		Long: `NexusAI is a next-generation decentralized AI agent platform that seamlessly 
integrates blockchain intelligence, cross-chain operations, and autonomous AI decision-making.

Built with 100% Golang and Bubble Tea TUI framework.`,
		Run: func(cmd *cobra.Command, args []string) {
			// Print cyberpunk banner
			printCyberpunkBanner()

			// Initialize the cyberpunk TUI application
			p := tea.NewProgram(
				tui.NewCyberpunkModel(),
				tea.WithAltScreen(),       // Full-screen TUI
				tea.WithMouseCellMotion(), // Mouse support
				tea.WithFPS(60),           // 60fps rendering
			)

			if _, err := p.Run(); err != nil {
				fmt.Printf("SYSTEM ERROR: %v\n", err)
				os.Exit(1)
			}
		},
	}

	// Add subcommands
	rootCmd.AddCommand(
		newWalletCommand(),
		newMarketCommand(),
		newAgentCommand(),
		newConfigCommand(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func newWalletCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "wallet",
		Short: "Manage wallets and view balances",
		Run: func(cmd *cobra.Command, args []string) {
			// Launch wallet-specific TUI
			fmt.Println("Launching wallet manager...")
		},
	}
}

func newMarketCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "market",
		Short: "View DeFi market data and analytics",
		Run: func(cmd *cobra.Command, args []string) {
			// Launch market-specific TUI
			fmt.Println("Launching market viewer...")
		},
	}
}

func newAgentCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "agent",
		Short: "Manage AI agents and automation",
		Run: func(cmd *cobra.Command, args []string) {
			// Launch agent-specific TUI
			fmt.Println("Launching agent manager...")
		},
	}
}

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Manage application configuration",
		Run: func(cmd *cobra.Command, args []string) {
			// Show configuration
			fmt.Println("Current configuration:")
			// TODO: Load and display config
		},
	}
}

// Print cyberpunk startup banner
func printCyberpunkBanner() {
	bannerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true).
		Padding(1, 2)

	banner := `
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

╔════════════════════════════════════════════════════════════════════════╗
║                    AEGIS PROTOCOL - CYBERPUNK EDITION                 ║
║              Intelligent Cross-Chain DeFi Terminal v2.0              ║
║                                                                      ║
║  FEATURES:                                                           ║
║  • Real-time Market Surveillance                                     ║
║  • Multi-Agent Coordination Matrix                                   ║
║  • Neural Interface for AI Communication                            ║
║  • Quantum-Encrypted Wallet Management                               ║
║  • Cross-Chain Protocol Integration                                  ║
║                                                                      ║
║  SPONSORS: Pyth Network • Blockscout • Hedera • Envio • Lit Protocol ║
╚════════════════════════════════════════════════════════════════════════╝`

	fmt.Println(bannerStyle.Render(banner))
	fmt.Println("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Render("Initializing cyberpunk interface..."))
}
