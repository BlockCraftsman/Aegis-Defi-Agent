package defi

import (
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// AaveManager handles Aave protocol interactions
type AaveManager struct {
	ContractManager *ContractManager
}

// NewAaveManager creates a new Aave manager
func NewAaveManager(cm *ContractManager) *AaveManager {
	return &AaveManager{
		ContractManager: cm,
	}
}

// DepositParams contains parameters for Aave deposit
type DepositParams struct {
	Asset        common.Address
	Amount       *big.Int
	OnBehalfOf   common.Address
	ReferralCode uint16
}

// WithdrawParams contains parameters for Aave withdraw
type WithdrawParams struct {
	Asset  common.Address
	Amount *big.Int
	To     common.Address
}

// BorrowParams contains parameters for Aave borrow
type BorrowParams struct {
	Asset            common.Address
	Amount           *big.Int
	InterestRateMode *big.Int
	ReferralCode     uint16
	OnBehalfOf       common.Address
}

// ExecuteDeposit deposits assets to Aave
func (am *AaveManager) ExecuteDeposit(params DepositParams) (*types.Transaction, error) {
	log.Printf("Executing Aave deposit: %s %s", params.Amount.String(), params.Asset.Hex())

	tx, err := am.ContractManager.TransactContract(
		"aave_lending_pool",
		AaveDeposit,
		big.NewInt(0),
		params.Asset,
		params.Amount,
		params.OnBehalfOf,
		params.ReferralCode,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute deposit: %v", err)
	}

	log.Printf("Aave deposit transaction sent: %s", tx.Hash().Hex())
	return tx, nil
}

// ExecuteWithdraw withdraws assets from Aave
func (am *AaveManager) ExecuteWithdraw(params WithdrawParams) (*types.Transaction, error) {
	log.Printf("Executing Aave withdraw: %s %s", params.Amount.String(), params.Asset.Hex())

	tx, err := am.ContractManager.TransactContract(
		"aave_lending_pool",
		AaveWithdraw,
		big.NewInt(0),
		params.Asset,
		params.Amount,
		params.To,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute withdraw: %v", err)
	}

	log.Printf("Aave withdraw transaction sent: %s", tx.Hash().Hex())
	return tx, nil
}

// ExecuteBorrow borrows assets from Aave
func (am *AaveManager) ExecuteBorrow(params BorrowParams) (*types.Transaction, error) {
	log.Printf("Executing Aave borrow: %s %s", params.Amount.String(), params.Asset.Hex())

	tx, err := am.ContractManager.TransactContract(
		"aave_lending_pool",
		AaveBorrow,
		big.NewInt(0),
		params.Asset,
		params.Amount,
		params.InterestRateMode,
		params.ReferralCode,
		params.OnBehalfOf,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute borrow: %v", err)
	}

	log.Printf("Aave borrow transaction sent: %s", tx.Hash().Hex())
	return tx, nil
}

// Common Aave interest rate modes
var (
	StableRate   = big.NewInt(1)
	VariableRate = big.NewInt(2)
)

// Common Aave referral code
const AaveReferralCode = 0

// GetSupportedAssets returns common Aave-supported assets
func (am *AaveManager) GetSupportedAssets() map[string]common.Address {
	return map[string]common.Address{
		"USDC": common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		"USDT": common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"),
		"DAI":  common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),
		"WETH": common.HexToAddress("0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"),
		"WBTC": common.HexToAddress("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599"),
		"LINK": common.HexToAddress("0x514910771AF9Ca656af840dff83E8264EcF986CA"),
	}
}

// MonitorTransaction monitors an Aave transaction for completion
func (am *AaveManager) MonitorTransaction(txHash common.Hash) (*types.Receipt, error) {
	log.Printf("Monitoring Aave transaction: %s", txHash.Hex())

	receipt, err := am.ContractManager.WaitForTransaction(txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to monitor transaction: %v", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		log.Printf("Aave transaction completed successfully in block %d", receipt.BlockNumber.Uint64())
	} else {
		log.Printf("Aave transaction failed in block %d", receipt.BlockNumber.Uint64())
	}

	return receipt, nil
}

// GetHealthFactor calculates health factor (simulated)
func (am *AaveManager) GetHealthFactor(user common.Address) (*big.Float, error) {
	// In a real implementation, this would query Aave's data provider
	// For now, return a simulated health factor

	// Simulate a healthy position
	healthFactor := big.NewFloat(2.5)
	log.Printf("Health factor for %s: %.2f", user.Hex(), healthFactor)

	return healthFactor, nil
}
