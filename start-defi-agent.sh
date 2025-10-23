#!/bin/bash

# Aegis Protocol - DeFi AI Agent Terminal Launcher

echo ""
echo "🚀 AEGIS PROTOCOL - DEFI AI AGENT TERMINAL"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Check if terminal is built
if [ ! -f "bin/aegis-terminal" ]; then
    echo -e "${YELLOW}⚠ Terminal not built. Building now...${NC}"
    make build-terminal
    if [ $? -ne 0 ]; then
        echo -e "${RED}✗ Failed to build terminal${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}✓ DeFi AI Agent Terminal ready${NC}"
echo ""
echo -e "${CYAN}🤖 Features:${NC}"
echo "  • Real-time DeFi market monitoring"
echo "  • AI-powered arbitrage detection"
echo "  • Multi-agent coordination"
echo "  • Portfolio management"
echo "  • Strategy automation"
echo ""
echo -e "${CYAN}🎮 Controls:${NC}"
echo "  • Use number keys 1-5 to navigate views"
echo "  • Type commands to interact with AI Agent"
echo "  • Press 'q' or Ctrl+C to exit"
echo ""
echo -e "${CYAN}📊 Available Views:${NC}"
echo "  1. Dashboard - System overview"
echo "  2. Agents - AI agent status"
echo "  3. Market - Real-time market data"
echo "  4. Wallet - Portfolio management"
echo "  5. Strategies - Trading strategies"
echo ""
echo -e "${CYAN}💬 Sample Commands:${NC}"
echo "  • 'deploy arbitrage' - Start arbitrage detection"
echo "  • 'check balance' - Show wallet balances"
echo "  • 'market status' - Current market conditions"
echo "  • 'agent status' - Agent performance"
echo "  • 'help' - Show all commands"
echo ""

read -p "Press Enter to start the DeFi AI Agent Terminal..."

# Launch the terminal
./bin/aegis-terminal