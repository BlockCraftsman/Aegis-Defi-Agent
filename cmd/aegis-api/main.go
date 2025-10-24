package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/api"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/monitoring"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/portfolio"
	"github.com/spf13/cobra"
)

var (
	port    int
	cfgFile string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aegis-api",
		Short: "Aegis DeFi Agent API Server",
		Long: `Aegis DeFi Agent API Server provides RESTful API access to
portfolio management, market data, DeFi strategies, and AI agents.

This server exposes the OpenAPI specification and handles all API requests.`,
		Run: runServer,
	}

	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	rootCmd.Flags().StringVarP(&cfgFile, "config", "c", "config/config.yaml", "Configuration file path")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Create a context that will be canceled on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	// Load configuration
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logging
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize monitoring
	monitor := monitoring.NewMonitor(&cfg.Monitoring, logger)

	// Initialize portfolio manager
	portfolioManager := portfolio.NewPortfolioManager(logger, monitor)

	// Create API server
	apiServer := api.NewServer(cfg, logger, monitor, portfolioManager)

	// Start API server
	logger.Info("Starting Aegis API server",
		logging.WithInt("port", port),
	)

	if err := apiServer.Start(port); err != nil {
		logger.Error("Failed to start API server",
			logging.WithError(err),
		)
		os.Exit(1)
	}

	// Print startup message
	printStartupMessage(port)

	// Wait for context cancellation
	<-ctx.Done()
	logger.Info("Shutting down API server")

	// Gracefully stop the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := apiServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error during API server shutdown",
			logging.WithError(err),
		)
	}

	logger.Info("API server shutdown complete")
}

func printStartupMessage(port int) {
	fmt.Printf(`
╔════════════════════════════════════════════════════════════════════════╗
║                    AEGIS DEFI AGENT API SERVER                        ║
║                                                                        ║
║  Server started successfully!                                          ║
║                                                                        ║
║  API Documentation: http://localhost:%d/api/docs                      ║
║  OpenAPI Spec:      http://localhost:%d/api/openapi.yaml              ║
║  Health Check:      http://localhost:%d/health                        ║
║  Metrics:           http://localhost:%d/metrics                       ║
║                                                                        ║
║  Available Endpoints:                                                  ║
║  • /api/v1/portfolio     - Portfolio management                        ║
║  • /api/v1/market/data   - Market data                                 ║
║  • /api/v1/defi/strategies - DeFi strategies                          ║
║  • /api/v1/agents        - AI agent management                         ║
║                                                                        ║
║  Press Ctrl+C to stop the server                                       ║
╚════════════════════════════════════════════════════════════════════════╝

`, port, port, port, port)
}
