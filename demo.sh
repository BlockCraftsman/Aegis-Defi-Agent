#!/bin/bash

# Aegis Protocol Demo Script
# This script demonstrates the core features of Aegis Protocol

echo "🚀 Aegis Protocol Demo - AI-Powered DeFi Agent"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}▶${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.24 or later."
    exit 1
fi

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
    print_error "Build failed. Please check the errors above."
    exit 1
fi

# Verify builds
print_step "Verifying builds..."
if [ -f "bin/aegis-mcp-client" ] && [ -f "bin/aegis-mcp-server" ]; then
    print_success "Build verification passed!"
    echo "  - bin/aegis-mcp-client ✓"
    echo "  - bin/aegis-mcp-server ✓"
else
    print_error "Build verification failed!"
    exit 1
fi

echo ""
print_step "Step 2: Configuration Setup"
echo "---------------------------------"

# Check if wallet password is set
if [ -z "$WALLET_PASSWORD" ]; then
    print_warning "WALLET_PASSWORD environment variable not set."
    echo "Setting demo password..."
    export WALLET_PASSWORD="demo-password-123"
    print_success "Demo password set: demo-password-123"
    echo "Note: For production, use a strong password and set WALLET_PASSWORD in your environment."
else
    print_success "Wallet password is configured"
fi

echo ""
print_step "Step 3: Starting Aegis MCP Server"
echo "---------------------------------------"

# Start server in background
print_step "Starting MCP server on port 18080..."
./bin/aegis-mcp-server &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Check if server is running
if ps -p $SERVER_PID > /dev/null; then
    print_success "MCP server started successfully (PID: $SERVER_PID)"
else
    print_error "Failed to start MCP server"
    exit 1
fi

echo ""
print_step "Step 4: Testing MCP Client Interaction"
echo "--------------------------------------------"

# Check if DeepSeek API key is available
if [ -z "$DEEPSEEK_KEY" ]; then
    print_warning "DEEPSEEK_KEY environment variable not set."
    echo ""
    echo "To test the full AI interaction, you need a DeepSeek API key:"
    echo "1. Get an API key from https://platform.deepseek.com/"
    echo "2. Set it as: export DEEPSEEK_KEY='your-api-key-here'"
    echo ""
    echo "For this demo, we'll show the server startup and architecture."
    echo ""
else
    print_success "DeepSeek API key is configured"
    
    # Create a test script for client interaction
    cat > /tmp/aegis_test_commands.txt << 'EOF'
Could you please greet my friend Alice for me?

What tools are available?

Can you help me with some data validation?
EOF

    print_step "Running client with test commands..."

    # Run client with test commands (non-interactive for demo)
    print_step "Testing 'say_hello' tool..."
    echo "Request: Could you please greet my friend Alice for me?"
    echo -n "Response: "
    ./bin/aegis-mcp-client -http http://localhost:18080/ -deepseek-key "$DEEPSEEK_KEY" -command "Could you please greet my friend Alice for me?" 2>/dev/null | grep -A 5 "AI Response:" | tail -n +2

    echo ""
    print_step "Testing tool discovery..."
    echo "Request: What tools are available?"
    echo -n "Response: "
    ./bin/aegis-mcp-client -http http://localhost:18080/ -deepseek-key "$DEEPSEEK_KEY" -command "What tools are available?" 2>/dev/null | grep -A 10 "AI Response:" | tail -n +2

    echo ""
    print_step "Step 5: Testing WASM Module Integration"
    echo "---------------------------------------------"

    # Test data validation WASM module
    print_step "Testing data validation tool..."
    TEST_JSON='{"data": "test", "signature": "abc123"}'
    echo "Request: Validate this JSON data: $TEST_JSON"
    echo -n "Response: "
    ./bin/aegis-mcp-client -http http://localhost:18080/ -deepseek-key "$DEEPSEEK_KEY" -command "Validate this JSON data: $TEST_JSON" 2>/dev/null | grep -A 5 "AI Response:" | tail -n +2
fi

echo ""
print_step "Step 6: Architecture Overview"
echo "-----------------------------------"

echo "Aegis Protocol Architecture:"
echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│                    Aegis Protocol                           │"
echo "│                 AI-Powered DeFi Agent                      │"
echo "├─────────────────────────────────────────────────────────────┤"
echo "│  Components Demonstrated:                                  │"
echo "│  • MCP Server (Port 18080)                                 │"
echo "│  • MCP Client                                              │"
echo "│  • WASM Module Integration                                 │"
echo "│  • IPFS Support                                            │"
echo "│  • AI Tool Discovery                                       │"
echo "│  • Real-time Data Processing                               │"
echo "└─────────────────────────────────────────────────────────────┘"
echo ""
echo "Supported DeFi Integrations:"
echo "  • Pyth Network (Price Feeds)"
echo "  • Hedera (Agent Coordination)"
echo "  • Lit Protocol (Automation)"
echo "  • Blockscout (Monitoring)"
echo "  • Envio (Data Indexing)"

echo ""
print_step "Step 7: Cleanup"
echo "-------------------"

# Stop the server
print_step "Stopping MCP server..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
print_success "MCP server stopped"

# Clean up test files
rm -f /tmp/aegis_test_commands.txt

echo ""
echo "🎉 Demo Completed Successfully!"
echo ""
echo "Next Steps:"
echo "1. Review the configuration in config/mcp_manifest.yaml"
echo "2. Add your own WASM modules to wasm-examples/"
echo "3. Configure real DeFi integrations with API keys"
echo "4. Explore the multi-agent coordination features"
echo ""
echo "For more information, see the README.md file."