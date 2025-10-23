#!/bin/bash

# Aegis Protocol Demo Verification Script
# This script verifies that all demo components are working correctly

set -e

echo ""
echo "🔍 AEGIS PROTOCOL - DEMO VERIFICATION"
echo "====================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Test functions
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected="$3"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    echo -n "Testing: $test_name... "
    
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

check_file() {
    local file="$1"
    local description="$2"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    echo -n "Checking: $description... "
    
    if [ -f "$file" ]; then
        echo -e "${GREEN}EXISTS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}MISSING${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

check_directory() {
    local dir="$1"
    local description="$2"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    
    echo -n "Checking: $description... "
    
    if [ -d "$dir" ]; then
        echo -e "${GREEN}EXISTS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}MISSING${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Phase 1: Project Structure Verification
echo "📁 PHASE 1: PROJECT STRUCTURE"
echo "-----------------------------"

check_file "go.mod" "Go module configuration"
check_file "Makefile" "Build automation"
check_file "README.md" "Documentation"
check_directory "cmd/" "Command-line applications"
check_directory "core/" "Core protocol implementation"
check_directory "internal/" "Internal packages"
check_directory "pkg/" "Public packages"
check_directory "wasm-examples/" "WASM module examples"
check_directory "config/" "Configuration files"
check_file "config/mcp_manifest.yaml" "MCP server configuration"

echo ""

# Phase 2: Build System Verification
echo "🔨 PHASE 2: BUILD SYSTEM"
echo "-------------------------"

run_test "Go installation" "go version" "Go compiler"
run_test "Project dependencies" "go mod download" "Dependency resolution"
run_test "Code compilation" "go build ./..." "Source code compilation"

# Build specific components
echo -n "Building: MCP Server... "
if go build -o /tmp/test-mcp-server ./cmd/aegis-mcp-server > /dev/null 2>&1; then
    echo -e "${GREEN}SUCCESS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo -n "Building: MCP Client... "
if go build -o /tmp/test-mcp-client ./cmd/aegis-mcp-client > /dev/null 2>&1; then
    echo -e "${GREEN}SUCCESS${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo ""

# Phase 3: Configuration Verification
echo "⚙️  PHASE 3: CONFIGURATION"
echo "--------------------------"

# Check configuration files
check_file "config/mcp_manifest.yaml" "MCP server manifest"

# Validate YAML configuration
echo -n "Validating: MCP manifest YAML... "
if python3 -c "import yaml; yaml.safe_load(open('config/mcp_manifest.yaml'))" > /dev/null 2>&1; then
    echo -e "${GREEN}VALID${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}INVALID${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

# Check WASM modules
echo -n "Checking: WASM module examples... "
WASM_COUNT=$(find wasm-examples/ -name "*.wasm" | wc -l)
if [ "$WASM_COUNT" -ge 2 ]; then
    echo -e "${GREEN}$WASM_COUNT modules found${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}Only $WASM_COUNT modules found${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo ""

# Phase 4: Core Components Verification
echo "🔧 PHASE 4: CORE COMPONENTS"
echo "---------------------------"

# Check core packages
check_directory "core/mcp/" "MCP protocol implementation"
check_file "core/mcp/mcp.go" "MCP core logic"
check_file "core/mcp/wallet.go" "Wallet management"
check_file "core/mcp/wasm.go" "WASM integration"

# Check DeFi components
check_directory "internal/defi/" "DeFi agent system"
check_file "internal/defi/agent.go" "DeFi agent implementation"
check_file "internal/defi/strategies.go" "Trading strategies"
check_file "internal/defi/contracts.go" "Smart contract interactions"

# Check MCP client integrations
check_directory "pkg/mcpclient/" "MCP client implementations"
check_file "pkg/mcpclient/client.go" "Base MCP client"
check_file "pkg/mcpclient/tools.go" "Tool definitions"

echo ""

# Phase 5: Demo Scripts Verification
echo "🎬 PHASE 5: DEMO SCRIPTS"
echo "------------------------"

check_file "demo-presentation.sh" "Presentation script"
check_file "demo.sh" "Full demo script"
check_file "demo-simple.sh" "Simple demo script"

# Make demo scripts executable
echo -n "Setting: Demo scripts executable... "
if chmod +x demo-presentation.sh demo.sh demo-simple.sh > /dev/null 2>&1; then
    echo -e "${GREEN}DONE${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}FAILED${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo ""

# Phase 6: Documentation Verification
echo "📚 PHASE 6: DOCUMENTATION"
echo "-------------------------"

check_file "docs/lightpaper.md" "Technical lightpaper"
check_file "docs/TECHNICAL_ARCHITECTURE.md" "Architecture documentation"
check_file "WASM_MODULE_DEVELOPMENT_AND_COMPILATION_GUIDE.md" "WASM development guide"

# Check if lightpaper has Mermaid diagrams
echo -n "Checking: Lightpaper Mermaid diagrams... "
if grep -q "```mermaid" docs/lightpaper.md; then
    echo -e "${GREEN}PRESENT${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
else
    echo -e "${RED}MISSING${NC}"
    TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_TOTAL=$((TESTS_TOTAL + 1))

echo ""

# Summary
echo "📊 VERIFICATION SUMMARY"
echo "======================="
echo ""
echo "Total Tests: $TESTS_TOTAL"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

SUCCESS_RATE=$((TESTS_PASSED * 100 / TESTS_TOTAL))
echo "Success Rate: $SUCCESS_RATE%"
echo ""

if [ "$TESTS_FAILED" -eq 0 ]; then
    echo -e "${GREEN}🎉 ALL TESTS PASSED! The demo is ready for presentation.${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Run: ./demo-presentation.sh (for full presentation)"
    echo "2. Run: ./demo-simple.sh (for quick overview)"
    echo "3. Run: ./demo.sh (for complete system demo)"
    exit 0
elif [ "$SUCCESS_RATE" -ge 80 ]; then
    echo -e "${YELLOW}⚠️  MOST TESTS PASSED! The demo is mostly ready.${NC}"
    echo ""
    echo "Some components need attention. Review the failed tests above."
    exit 1
else
    echo -e "${RED}❌ SIGNIFICANT ISSUES DETECTED! The demo needs work.${NC}"
    echo ""
    echo "Please address the failed tests before presenting."
    exit 1
fi