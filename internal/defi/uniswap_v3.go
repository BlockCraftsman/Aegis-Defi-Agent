package defi

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// UniswapV3Manager handles Uniswap V3 specific operations
type UniswapV3Manager struct {
	ContractManager *ContractManager
}

// NewUniswapV3Manager creates a new Uniswap V3 manager
func NewUniswapV3Manager(cm *ContractManager) *UniswapV3Manager {
	return &UniswapV3Manager{
		ContractManager: cm,
	}
}

// SwapParams contains parameters for a Uniswap V3 swap
type SwapParams struct {
	TokenIn           common.Address
	TokenOut          common.Address
	Fee               uint24
	Recipient         common.Address
	Deadline          *big.Int
	AmountIn          *big.Int
	AmountOutMinimum  *big.Int
	SqrtPriceLimitX96 *big.Int
}

// ExecuteSwap executes a single token swap on Uniswap V3
func (uv3 *UniswapV3Manager) ExecuteSwap(params SwapParams) (*types.Transaction, error) {
	log.Printf("Executing Uniswap V3 swap: %s -> %s (amount: %s)",
		params.TokenIn.Hex(), params.TokenOut.Hex(), params.AmountIn.String())

	tx, err := uv3.ContractManager.TransactContract(
		"uniswap_v3_router",
		UniswapV3ExactInputSingle,
		big.NewInt(0),
		params.TokenIn,
		params.TokenOut,
		params.Fee,
		params.Recipient,
		params.Deadline,
		params.AmountIn,
		params.AmountOutMinimum,
		params.SqrtPriceLimitX96,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute swap: %v", err)
	}

	log.Printf("Uniswap V3 swap transaction sent: %s", tx.Hash().Hex())
	return tx, nil
}

// GetSwapQuote gets a quote for a Uniswap V3 swap (simulated)
func (uv3 *UniswapV3Manager) GetSwapQuote(params SwapParams) (*big.Int, error) {
	// In a real implementation, this would query the Uniswap V3 Quoter contract
	// For now, we'll simulate a reasonable output amount

	// Simple simulation: assume 0.3% fee and some slippage
	feeMultiplier := big.NewFloat(0.997) // 0.3% fee
	amountInFloat := new(big.Float).SetInt(params.AmountIn)

	// Apply fee
	amountOutFloat := new(big.Float).Mul(amountInFloat, feeMultiplier)

	// Convert back to big.Int
	amountOut := new(big.Int)
	amountOutFloat.Int(amountOut)

	log.Printf("Swap quote: %s %s -> %s %s (estimated)",
		params.AmountIn.String(), params.TokenIn.Hex(),
		amountOut.String(), params.TokenOut.Hex())

	return amountOut, nil
}

// Common Uniswap V3 fee tiers
const (
	FeeTierLow    uint24 = 500   // 0.05%
	FeeTierMedium uint24 = 3000  // 0.3%
	FeeTierHigh   uint24 = 10000 // 1%
)

// GetOptimalFeeTier determines the optimal fee tier for a token pair
func (uv3 *UniswapV3Manager) GetOptimalFeeTier(tokenA, tokenB common.Address) uint24 {
	// Common stablecoin pairs use lower fees
	stablecoins := map[common.Address]bool{
		common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"): true, // USDC
		common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"): true, // USDT
		common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"): true, // DAI
	}

	if stablecoins[tokenA] && stablecoins[tokenB] {
		return FeeTierLow
	}

	// Major tokens with high liquidity
	majorTokens := map[common.Address]bool{
		common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"): true, // WETH
		common.HexToAddress("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599"): true, // WBTC
	}

	if majorTokens[tokenA] || majorTokens[tokenB] {
		return FeeTierMedium
	}

	// Default to higher fee for less liquid pairs
	return FeeTierHigh
}

// CreateDeadline creates a deadline timestamp for transactions
func CreateDeadline(minutes int64) *big.Int {
	deadline := big.NewInt(time.Now().Unix() + minutes*60)
	return deadline
}

// MonitorSwap monitors a swap transaction for completion
func (uv3 *UniswapV3Manager) MonitorSwap(txHash common.Hash) (*types.Receipt, error) {
	log.Printf("Monitoring swap transaction: %s", txHash.Hex())

	receipt, err := uv3.ContractManager.WaitForTransaction(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor swap: %v", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Printf("Swap completed successfully in block %d", receipt.BlockNumber.Uint64())
	} else {
		log.Printf("Swap failed in block %d", receipt.BlockNumber.Uint64())
	}

	return receipt, nil
}
