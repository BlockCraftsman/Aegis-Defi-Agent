package defi

import (
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ChainConfig contains configuration for a blockchain network
type ChainConfig struct {
	Name        string
	ChainID     *big.Int
	RPCURL      string
	ExplorerURL string
	NativeToken string
	IsTestnet   bool
}

// MultiChainManager handles interactions with multiple blockchain networks
type MultiChainManager struct {
	Chains  map[string]*ChainConfig
	Clients map[string]*ethclient.Client
}

// NewMultiChainManager creates a new multi-chain manager
func NewMultiChainManager() *MultiChainManager {
	mcm := &MultiChainManager{
		Chains:  make(map[string]*ChainConfig),
		Clients: make(map[string]*ethclient.Client),
	}

	// Initialize common chains
	mcm.initializeCommonChains()

	return mcm
}

// initializeCommonChains initializes common blockchain networks
func (mcm *MultiChainManager) initializeCommonChains() {
	// Ethereum Mainnet
	mcm.Chains["ethereum"] = &ChainConfig{
		Name:        "Ethereum",
		ChainID:     big.NewInt(1),
		RPCURL:      getEnvWithFallback("ETH_RPC_URL", "https://eth-mainnet.g.alchemy.com/v2/demo"),
		ExplorerURL: "https://etherscan.io",
		NativeToken: "ETH",
		IsTestnet:   false,
	}

	// Polygon Mainnet
	mcm.Chains["polygon"] = &ChainConfig{
		Name:        "Polygon",
		ChainID:     big.NewInt(137),
		RPCURL:      getEnvWithFallback("POLYGON_RPC_URL", "https://polygon-mainnet.g.alchemy.com/v2/demo"),
		ExplorerURL: "https://polygonscan.com",
		NativeToken: "MATIC",
		IsTestnet:   false,
	}

	// Arbitrum Mainnet
	mcm.Chains["arbitrum"] = &ChainConfig{
		Name:        "Arbitrum",
		ChainID:     big.NewInt(42161),
		RPCURL:      getEnvWithFallback("ARBITRUM_RPC_URL", "https://arb-mainnet.g.alchemy.com/v2/demo"),
		ExplorerURL: "https://arbiscan.io",
		NativeToken: "ETH",
		IsTestnet:   false,
	}

	// Optimism Mainnet
	mcm.Chains["optimism"] = &ChainConfig{
		Name:        "Optimism",
		ChainID:     big.NewInt(10),
		RPCURL:      getEnvWithFallback("OPTIMISM_RPC_URL", "https://opt-mainnet.g.alchemy.com/v2/demo"),
		ExplorerURL: "https://optimistic.etherscan.io",
		NativeToken: "ETH",
		IsTestnet:   false,
	}

	// Testnets
	mcm.Chains["goerli"] = &ChainConfig{
		Name:        "Goerli",
		ChainID:     big.NewInt(5),
		RPCURL:      getEnvWithFallback("GOERLI_RPC_URL", "https://eth-goerli.g.alchemy.com/v2/demo"),
		ExplorerURL: "https://goerli.etherscan.io",
		NativeToken: "ETH",
		IsTestnet:   true,
	}

	mcm.Chains["polygon_mumbai"] = &ChainConfig{
		Name:        "Polygon Mumbai",
		ChainID:     big.NewInt(80001),
		RPCURL:      getEnvWithFallback("MUMBAI_RPC_URL", "https://polygon-mumbai.g.alchemy.com/v2/demo"),
		ExplorerURL: "https://mumbai.polygonscan.com",
		NativeToken: "MATIC",
		IsTestnet:   true,
	}
}

// ConnectToChain connects to a specific blockchain network
func (mcm *MultiChainManager) ConnectToChain(chainName string) (*ethclient.Client, error) {
	config, exists := mcm.Chains[chainName]
	if !exists {
		return nil, fmt.Errorf("chain %s not supported", chainName)
	}

	// Check if we already have a connection
	if client, exists := mcm.Clients[chainName]; exists {
		return client, nil
	}

	// Create new connection
	client, err := ethclient.Dial(config.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", chainName, err)
	}

	// Verify connection by getting chain ID
	chainID, err := client.ChainID(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID for %s: %v", chainName, err)
	}

	if chainID.Cmp(config.ChainID) != 0 {
		return nil, fmt.Errorf("chain ID mismatch for %s: expected %s, got %s",
			chainName, config.ChainID.String(), chainID.String())
	}

	mcm.Clients[chainName] = client
	log.Printf("Connected to %s (Chain ID: %s)", config.Name, chainID.String())

	return client, nil
}

// GetChainConfig returns configuration for a chain
func (mcm *MultiChainManager) GetChainConfig(chainName string) (*ChainConfig, error) {
	config, exists := mcm.Chains[chainName]
	if !exists {
		return nil, fmt.Errorf("chain %s not found", chainName)
	}
	return config, nil
}

// ListSupportedChains returns list of supported chains
func (mcm *MultiChainManager) ListSupportedChains() []string {
	chains := make([]string, 0, len(mcm.Chains))
	for name := range mcm.Chains {
		chains = append(chains, name)
	}
	return chains
}

// GetContractAddresses returns contract addresses for a specific chain
func (mcm *MultiChainManager) GetContractAddresses(chainName string) (map[string]common.Address, error) {
	config, err := mcm.GetChainConfig(chainName)
	if err != nil {
		return nil, err
	}

	addresses := make(map[string]common.Address)

	switch chainName {
	case "ethereum":
		addresses["uniswap_v2_router"] = common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
		addresses["uniswap_v3_router"] = common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
		addresses["aave_lending_pool"] = common.HexToAddress("0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9")
		addresses["usdc"] = common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")
		addresses["usdt"] = common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
		addresses["dai"] = common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F")
		addresses["weth"] = common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2")

	case "polygon":
		addresses["uniswap_v2_router"] = common.HexToAddress("0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff")
		addresses["uniswap_v3_router"] = common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
		addresses["aave_lending_pool"] = common.HexToAddress("0x8dFf5E27EA6b7AC08EbFdf9eB090F32ee9a30fcf")
		addresses["usdc"] = common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
		addresses["usdt"] = common.HexToAddress("0xc2132D05D31c914a87C6611C10748AEb04B58e8F")
		addresses["dai"] = common.HexToAddress("0x8f3Cf7ad23Cd3CaDbD9735AFf958023239c6A063")
		addresses["weth"] = common.HexToAddress("0x7ceB23fD6bC0adD59E62ac25578270cFf1b9f619")

	case "arbitrum":
		addresses["uniswap_v3_router"] = common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")
		addresses["aave_lending_pool"] = common.HexToAddress("0x794a61358D6845594F94dc1DB02A252b5b4814aD")
		addresses["usdc"] = common.HexToAddress("0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8")
		addresses["usdt"] = common.HexToAddress("0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9")
		addresses["dai"] = common.HexToAddress("0xDA10009cBd5D07dd0CeCc66161FC93D7c9000da1")
		addresses["weth"] = common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1")

	default:
		return nil, fmt.Errorf("contract addresses not configured for %s", chainName)
	}

	log.Printf("Loaded %d contract addresses for %s", len(addresses), config.Name)
	return addresses, nil
}

// getEnvWithFallback gets environment variable with fallback
func getEnvWithFallback(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// IsChainSupported checks if a chain is supported
func (mcm *MultiChainManager) IsChainSupported(chainName string) bool {
	_, exists := mcm.Chains[chainName]
	return exists
}

// GetChainStatus returns status information for a chain
func (mcm *MultiChainManager) GetChainStatus(chainName string) (map[string]interface{}, error) {
	client, err := mcm.ConnectToChain(chainName)
	if err != nil {
		return nil, err
	}

	config := mcm.Chains[chainName]

	blockNumber, err := client.BlockNumber(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get block number: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	status := map[string]interface{}{
		"chain_name":   config.Name,
		"chain_id":     config.ChainID.String(),
		"block_number": blockNumber,
		"gas_price":    gasPrice.String(),
		"native_token": config.NativeToken,
		"is_testnet":   config.IsTestnet,
		"explorer_url": config.ExplorerURL,
	}

	return status, nil
}
