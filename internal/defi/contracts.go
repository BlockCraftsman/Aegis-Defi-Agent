package defi

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// uint24 type for Uniswap V3 fees
type uint24 uint32

// ContractManager handles interactions with DeFi smart contracts
type ContractManager struct {
	Client     *ethclient.Client
	Transactor *bind.TransactOpts
	Contracts  map[string]*DeFiContract
}

// DeFiContract represents a DeFi protocol contract
type DeFiContract struct {
	Name     string
	Address  common.Address
	ABI      abi.ABI
	Instance interface{}
}

// NewContractManager creates a new contract manager with real blockchain connection
func NewContractManager(client *ethclient.Client, privateKey string) (*ContractManager, error) {
	if client == nil {
		// Create a real blockchain connection
		rpcURL := getRPCURL()
		var err error
		client, err = ethclient.Dial(rpcURL)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to blockchain: %v", err)
		}
	}

	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	transactor, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	// Get current nonce and gas price
	address := crypto.PubkeyToAddress(key.PublicKey)
	nonce, err := client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	transactor.Nonce = big.NewInt(int64(nonce))
	transactor.GasPrice = gasPrice
	transactor.GasLimit = 300000

	manager := &ContractManager{
		Client:     client,
		Transactor: transactor,
		Contracts:  make(map[string]*DeFiContract),
	}

	// Initialize common DeFi contracts
	if err := manager.initializeCommonContracts(); err != nil {
		log.Printf("Warning: Failed to initialize some contracts: %v", err)
	}

	return manager, nil
}

// initializeCommonContracts initializes common DeFi contracts
func (cm *ContractManager) initializeCommonContracts() error {
	// Initialize Uniswap V2 Router
	uniswapV2RouterABI := `[
		{
			"inputs": [
				{"internalType": "uint256", "name": "amountIn", "type": "uint256"},
				{"internalType": "address[]", "name": "path", "type": "address[]"}
			],
			"name": "getAmountsOut",
			"outputs": [
				{"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
			],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [
				{"internalType": "uint256", "name": "amountOutMin", "type": "uint256"},
				{"internalType": "address[]", "name": "path", "type": "address[]"},
				{"internalType": "address", "name": "to", "type": "address"},
				{"internalType": "uint256", "name": "deadline", "type": "uint256"}
			],
			"name": "swapExactETHForTokens",
			"outputs": [
				{"internalType": "uint256[]", "name": "amounts", "type": "uint256[]"}
			],
			"stateMutability": "payable",
			"type": "function"
		}
	]`

	// Initialize Uniswap V3 Router
	uniswapV3RouterABI := `[
		{
			"inputs": [
				{
					"components": [
						{"internalType": "address", "name": "tokenIn", "type": "address"},
						{"internalType": "address", "name": "tokenOut", "type": "address"},
						{"internalType": "uint24", "name": "fee", "type": "uint24"},
						{"internalType": "address", "name": "recipient", "type": "address"},
						{"internalType": "uint256", "name": "deadline", "type": "uint256"},
						{"internalType": "uint256", "name": "amountIn", "type": "uint256"},
						{"internalType": "uint256", "name": "amountOutMinimum", "type": "uint256"},
						{"internalType": "uint160", "name": "sqrtPriceLimitX96", "type": "uint160"}
					],
					"internalType": "struct ISwapRouter.ExactInputSingleParams",
					"name": "params",
					"type": "tuple"
				}
			],
			"name": "exactInputSingle",
			"outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
			"stateMutability": "payable",
			"type": "function"
		},
		{
			"inputs": [
				{
					"components": [
						{"internalType": "bytes", "name": "path", "type": "bytes"},
						{"internalType": "address", "name": "recipient", "type": "address"},
						{"internalType": "uint256", "name": "deadline", "type": "uint256"},
						{"internalType": "uint256", "name": "amountIn", "type": "uint256"},
						{"internalType": "uint256", "name": "amountOutMinimum", "type": "uint256"}
					],
					"internalType": "struct ISwapRouter.ExactInputParams",
					"name": "params",
					"type": "tuple"
				}
			],
			"name": "exactInput",
			"outputs": [{"internalType": "uint256", "name": "amountOut", "type": "uint256"}],
			"stateMutability": "payable",
			"type": "function"
		}
	]`

	if err := cm.AddContract("uniswap_v2_router", UniswapV2Router, uniswapV2RouterABI); err != nil {
		return fmt.Errorf("failed to add Uniswap V2 Router: %v", err)
	}

	// Initialize Uniswap V3 Router
	if err := cm.AddContract("uniswap_v3_router", UniswapV3Router, uniswapV3RouterABI); err != nil {
		return fmt.Errorf("failed to add Uniswap V3 Router: %v", err)
	}

	// Initialize Aave Lending Pool
	aaveLendingPoolABI := `[
		{
			"inputs": [
				{"internalType": "address", "name": "asset", "type": "address"},
				{"internalType": "uint256", "name": "amount", "type": "uint256"},
				{"internalType": "address", "name": "onBehalfOf", "type": "address"},
				{"internalType": "uint16", "name": "referralCode", "type": "uint16"}
			],
			"name": "deposit",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"inputs": [
				{"internalType": "address", "name": "asset", "type": "address"},
				{"internalType": "uint256", "name": "amount", "type": "uint256"},
				{"internalType": "address", "name": "to", "type": "address"}
			],
			"name": "withdraw",
			"outputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
			"stateMutability": "nonpayable",
			"type": "function"
		},
		{
			"inputs": [
				{"internalType": "address", "name": "asset", "type": "address"},
				{"internalType": "uint256", "name": "amount", "type": "uint256"},
				{"internalType": "uint256", "name": "interestRateMode", "type": "uint256"},
				{"internalType": "uint16", "name": "referralCode", "type": "uint16"},
				{"internalType": "address", "name": "onBehalfOf", "type": "address"}
			],
			"name": "borrow",
			"outputs": [],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`

	if err := cm.AddContract("aave_lending_pool", AaveLendingPool, aaveLendingPoolABI); err != nil {
		return fmt.Errorf("failed to add Aave Lending Pool: %v", err)
	}

	// Initialize ERC20 token interface (for common tokens)
	erc20ABI := `[
		{
			"inputs": [{"internalType": "address", "name": "account", "type": "address"}],
			"name": "balanceOf",
			"outputs": [{"internalType": "uint256", "name": "", "type": "uint256"}],
			"stateMutability": "view",
			"type": "function"
		},
		{
			"inputs": [
				{"internalType": "address", "name": "spender", "type": "address"},
				{"internalType": "uint256", "name": "amount", "type": "uint256"}
			],
			"name": "approve",
			"outputs": [{"internalType": "bool", "name": "", "type": "bool"}],
			"stateMutability": "nonpayable",
			"type": "function"
		}
	]`

	// Add common ERC20 tokens
	commonTokens := map[string]common.Address{
		"USDC": common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		"USDT": common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
		"DAI":  common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
		"WETH": common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
	}

	for name, address := range commonTokens {
		if err := cm.AddContract("erc20_"+name, address, erc20ABI); err != nil {
			log.Printf("Warning: Failed to add %s token: %v", name, err)
		}
	}

	log.Printf("Initialized %d common contracts", len(cm.Contracts))
	return nil
}

// getRPCURL returns the appropriate RPC URL based on environment
func getRPCURL() string {
	// Check environment variables first
	if url := os.Getenv("ETH_RPC_URL"); url != "" {
		return url
	}
	if url := os.Getenv("POLYGON_RPC_URL"); url != "" {
		return url
	}
	if url := os.Getenv("ARBITRUM_RPC_URL"); url != "" {
		return url
	}

	// Fallback to public RPC endpoints
	return "https://eth-mainnet.g.alchemy.com/v2/demo"
}

// AddContract adds a new contract to the manager
func (cm *ContractManager) AddContract(name string, address common.Address, abiJSON string) error {
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return fmt.Errorf("failed to parse ABI for %s: %v", name, err)
	}

	cm.Contracts[name] = &DeFiContract{
		Name:    name,
		Address: address,
		ABI:     contractABI,
	}

	log.Printf("Added contract: %s at %s", name, address.Hex())
	return nil
}

// CallContract calls a contract method (read-only)
func (cm *ContractManager) CallContract(contractName, method string, args ...interface{}) ([]interface{}, error) {
	contract, exists := cm.Contracts[contractName]
	if !exists {
		return nil, fmt.Errorf("contract %s not found", contractName)
	}

	// Pack the method call
	data, err := contract.ABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method %s: %v", method, err)
	}

	// Call the contract
	result, err := cm.Client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &contract.Address,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	// Unpack the result
	var unpacked []interface{}
	if len(result) > 0 {
		unpacked, err = contract.ABI.Unpack(method, result)
		if err != nil {
			return nil, fmt.Errorf("failed to unpack result: %v", err)
		}
	}

	return unpacked, nil
}

// TransactContract sends a transaction to a contract
func (cm *ContractManager) TransactContract(contractName, method string, value *big.Int, args ...interface{}) (*types.Transaction, error) {
	contract, exists := cm.Contracts[contractName]
	if !exists {
		return nil, fmt.Errorf("contract %s not found", contractName)
	}

	// Pack the method call
	data, err := contract.ABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to pack method %s: %v", method, err)
	}

	// Estimate gas
	gasLimit, err := cm.Client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  cm.Transactor.From,
		To:    &contract.Address,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %v", err)
	}

	// Get current gas price
	gasPrice, err := cm.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	// Create and send transaction
	tx := types.NewTransaction(
		cm.Transactor.Nonce.Uint64(),
		contract.Address,
		value,
		gasLimit,
		gasPrice,
		data,
	)

	// Sign the transaction
	signedTx, err := cm.Transactor.Signer(cm.Transactor.From, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send the transaction
	err = cm.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	// Increment nonce
	cm.Transactor.Nonce.Add(cm.Transactor.Nonce, big.NewInt(1))

	log.Printf("Transaction sent: %s", signedTx.Hash().Hex())
	return signedTx, nil
}

// GetTransactionReceipt gets receipt for a transaction
func (cm *ContractManager) GetTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	receipt, err := cm.Client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %v", err)
	}
	return receipt, nil
}

// WaitForTransaction waits for transaction confirmation
func (cm *ContractManager) WaitForTransaction(txHash common.Hash) (*types.Receipt, error) {
	for {
		receipt, err := cm.GetTransactionReceipt(txHash)
		if err == nil {
			return receipt, nil
		}
		if err != ethereum.NotFound {
			return nil, err
		}
		// Wait before retrying
		time.Sleep(5 * time.Second)
	}
}

// Common DeFi Contract Addresses (Ethereum Mainnet)
var (
	// Uniswap V2
	UniswapV2Router  = common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	UniswapV2Factory = common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")

	// Uniswap V3
	UniswapV3Router = common.HexToAddress("0xE592427A0AEce92De3Edee1F18E0157C05861564")

	// Aave V2
	AaveLendingPool = common.HexToAddress("0x7d2768dE32b0b80b7a3454c06BdAc94A69DDc7A9")

	// Compound
	CompoundComptroller = common.HexToAddress("0x3d9819210A31b4961b30EF54bE2aeD79B9c9Cd3B")

	// Curve
	CurveRegistry = common.HexToAddress("0x90E00ACe148ca3b23Ac1bC8C240C2a7Dd9c2d7f5")
)

// Common ABI snippets for DeFi protocols
const (
	// Uniswap V2 Router ABI methods
	UniswapV2SwapExactETHForTokens = "swapExactETHForTokens"
	UniswapV2SwapExactTokensForETH = "swapExactTokensForETH"
	UniswapV2GetAmountsOut         = "getAmountsOut"

	// Uniswap V3 Router ABI methods
	UniswapV3ExactInputSingle = "exactInputSingle"
	UniswapV3ExactInput       = "exactInput"

	// Aave ABI methods
	AaveDeposit  = "deposit"
	AaveWithdraw = "withdraw"
	AaveBorrow   = "borrow"

	// ERC20 ABI methods
	ERC20BalanceOf = "balanceOf"
	ERC20Transfer  = "transfer"
	ERC20Approve   = "approve"
)

// Example: Swap ETH for tokens on Uniswap V2
func (cm *ContractManager) SwapETHForTokens(
	amountIn *big.Int,
	path []common.Address,
	to common.Address,
	deadline *big.Int,
	amountOutMin *big.Int,
) (*types.Transaction, error) {

	return cm.TransactContract(
		"uniswap_v2_router",
		UniswapV2SwapExactETHForTokens,
		amountIn,
		amountOutMin,
		path,
		to,
		deadline,
	)
}

// Example: Get expected output amount from Uniswap
func (cm *ContractManager) GetExpectedOutput(
	amountIn *big.Int,
	path []common.Address,
) (*big.Int, error) {

	result, err := cm.CallContract(
		"uniswap_v2_router",
		UniswapV2GetAmountsOut,
		amountIn,
		path,
	)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no result from getAmountsOut")
	}

	amounts, ok := result[0].([]*big.Int)
	if !ok || len(amounts) < 2 {
		return nil, fmt.Errorf("invalid result format")
	}

	return amounts[len(amounts)-1], nil
}

// UniswapV3ExactInputSingle executes a single token swap on Uniswap V3
func (cm *ContractManager) UniswapV3ExactInputSingle(
	tokenIn common.Address,
	tokenOut common.Address,
	fee uint24,
	recipient common.Address,
	deadline *big.Int,
	amountIn *big.Int,
	amountOutMinimum *big.Int,
	sqrtPriceLimitX96 *big.Int,
) (*types.Transaction, error) {

	// Create the params struct
	params := struct {
		TokenIn           common.Address
		TokenOut          common.Address
		Fee               uint24
		Recipient         common.Address
		Deadline          *big.Int
		AmountIn          *big.Int
		AmountOutMinimum  *big.Int
		SqrtPriceLimitX96 *big.Int
	}{
		TokenIn:           tokenIn,
		TokenOut:          tokenOut,
		Fee:               fee,
		Recipient:         recipient,
		Deadline:          deadline,
		AmountIn:          amountIn,
		AmountOutMinimum:  amountOutMinimum,
		SqrtPriceLimitX96: sqrtPriceLimitX96,
	}

	return cm.TransactContract(
		"uniswap_v3_router",
		UniswapV3ExactInputSingle,
		big.NewInt(0),
		params,
	)
}

// Example: Deposit to Aave
func (cm *ContractManager) AaveDeposit(
	asset common.Address,
	amount *big.Int,
	onBehalfOf common.Address,
	referralCode uint16,
) (*types.Transaction, error) {

	return cm.TransactContract(
		"aave_lending_pool",
		AaveDeposit,
		big.NewInt(0),
		asset,
		amount,
		onBehalfOf,
		referralCode,
	)
}

// GetTokenBalance gets balance of an ERC20 token
func (cm *ContractManager) GetTokenBalance(token, account common.Address) (*big.Int, error) {
	result, err := cm.CallContract(
		"erc20_"+token.Hex(),
		ERC20BalanceOf,
		account,
	)
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no balance returned")
	}

	balance, ok := result[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("invalid balance format")
	}

	return balance, nil
}
