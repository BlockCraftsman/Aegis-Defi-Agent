#!/bin/bash

# Aegis Protocol Simple Demo Script
# Shows the core architecture and components without requiring API keys

echo "🚀 Aegis Protocol Demo - AI-Powered DeFi Agent"
echo "================================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}▶${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_step "Step 1: Building Aegis Protocol Components"
echo "------------------------------------------------"

# Clean previous builds
print_step "Cleaning previous builds..."
make clean

# Build all components
print_step "Building Aegis Protocol components..."
if make all; then
    print_success "All components built successfully!"
else
    echo "Build failed. Please check the errors above."
    exit 1
fi

# Verify builds
print_step "Verifying builds..."
if [ -f "bin/aegis-mcp-client" ] && [ -f "bin/aegis-mcp-server" ]; then
    print_success "Build verification passed!"
    echo "  - bin/aegis-mcp-client ✓"
    echo "  - bin/aegis-mcp-server ✓"
else
    echo "Build verification failed!"
    exit 1
fi

echo ""
print_step "Step 2: Architecture Overview"
echo "-----------------------------------"

echo "Aegis Protocol Architecture:"
echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│                    Aegis Protocol                           │"
echo "│                 AI-Powered DeFi Agent                      │"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│  Core Components:                                           │"
echo "│  • MCP Server - Standardized AI tool communication         │"
echo "│  • MCP Client - Interface for AI interactions              │"
echo "│  • WASM Modules - Extensible functionality                 │"
echo "│  • IPFS Integration - Decentralized storage                │"
echo "│  • AI Agent Coordination - Multi-agent strategies          │"
echo "└─────────────────────────────────────────────────────────────┘"
echo ""

echo "DeFi Integrations:"
echo "  • Pyth Network - Real-time price feeds"
echo "  • Hedera - Agent discovery and coordination"
echo "  • Lit Protocol - Conditional automation"
echo "  • Blockscout - Transaction monitoring"
echo "  • Envio - Fast blockchain data indexing"

echo ""
print_step "Step 3: Configuration Overview"
echo "------------------------------------"

echo "Key Configuration Files:"
echo ""
echo "config/mcp_manifest.yaml - Main server configuration:"
echo "  • Server settings (host, port, timeout)"
echo "  • IPFS integration settings"
echo "  • LLM provider configuration"
echo "  • WASM module definitions"
echo ""

echo "Available WASM Modules:"
echo "  • hello - Greeting tool (loaded from IPFS)"
echo "  • data_validation - JSON validation tool"

echo ""
print_step "Step 4: Quick Start Commands"
echo "----------------------------------"

echo "To run the full system:"
echo ""
echo "1. Set wallet password:"
echo "   export WALLET_PASSWORD='your-password'"
echo ""
echo "2. Set DeepSeek API key (for AI interactions):"
echo "   export DEEPSEEK_KEY='your-api-key'"
echo ""
echo "3. Start MCP Server:"
echo "   ./bin/aegis-mcp-server"
echo ""
echo "4. Interact with Client:"
echo "   ./bin/aegis-mcp-client -http http://localhost:18080/ -deepseek-key \$DEEPSEEK_KEY"
echo ""

echo "Demo Commands to Try:"
echo "  • 'Could you please greet my friend Alice for me?'"
echo "  • 'What tools are available?'"
echo "  • 'Validate this JSON data: {\"data\": \"test\", \"signature\": \"abc123\"}'"

echo ""
print_step "Step 5: Development Workflow"
echo "----------------------------------"

echo "Development Commands:"
echo ""
echo "make run-server    # Run server directly (no build)"
echo "make run-client    # Run client directly (no build)"
echo "make build-all     # Build for all platforms"
echo "make clean         # Clean build artifacts"
echo ""

echo "Testing:"
echo "go test ./...      # Run all tests"
echo ""

print_success "Demo Overview Complete!"
echo ""
echo "Next Steps:"
echo "1. Review config/mcp_manifest.yaml for configuration options"
echo "2. Add your own WASM modules to wasm-examples/"
echo "3. Configure real DeFi integrations with API keys"
echo "4. Explore the multi-agent coordination features"
echo ""
echo "For detailed instructions, see the README.md file."