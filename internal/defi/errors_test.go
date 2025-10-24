package defi

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestDeFiError(t *testing.T) {
	t.Run("NewDeFiError", func(t *testing.T) {
		err := NewDeFiError(ErrValidation, "test error", map[string]interface{}{
			"field": "test_field",
		})

		assert.Equal(t, ErrValidation, err.Type)
		assert.Equal(t, "test error", err.Message)
		assert.Equal(t, "test_field", err.Details["field"])
		assert.Nil(t, err.Cause)
		assert.Contains(t, err.Error(), "VALIDATION_ERROR: test error")
	})

	t.Run("WrapDeFiError", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := WrapDeFiError(ErrBlockchain, "wrapped error", originalErr, map[string]interface{}{
			"operation": "test_op",
		})

		assert.Equal(t, ErrBlockchain, err.Type)
		assert.Equal(t, "wrapped error", err.Message)
		assert.Equal(t, "test_op", err.Details["operation"])
		assert.Equal(t, originalErr, err.Cause)
		assert.Contains(t, err.Error(), "BLOCKCHAIN_ERROR: wrapped error (cause: original error)")
	})

	t.Run("Unwrap", func(t *testing.T) {
		originalErr := errors.New("original error")
		err := WrapDeFiError(ErrStrategy, "test", originalErr, nil)

		unwrapped := errors.Unwrap(err)
		assert.Equal(t, originalErr, unwrapped)
	})
}

func TestValidateStrategy(t *testing.T) {
	t.Run("Valid Strategy", func(t *testing.T) {
		strategy := Strategy{
			Type: StrategyArbitrage,
			Conditions: []Condition{
				{
					Metric:    "price_difference",
					Operator:  ">",
					Threshold: 0.01,
				},
			},
			IsEnabled: true,
		}

		err := ValidateStrategy(strategy)
		assert.NoError(t, err)
	})

	t.Run("Empty Strategy Type", func(t *testing.T) {
		strategy := Strategy{
			Type: "",
		}

		err := ValidateStrategy(strategy)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "strategy type cannot be empty")
	})

	t.Run("Enabled Strategy Without Conditions", func(t *testing.T) {
		strategy := Strategy{
			Type:      StrategyArbitrage,
			IsEnabled: true,
		}

		err := ValidateStrategy(strategy)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "enabled strategy must have conditions")
	})

	t.Run("Invalid Condition Operator", func(t *testing.T) {
		strategy := Strategy{
			Type: StrategyArbitrage,
			Conditions: []Condition{
				{
					Metric:    "price_difference",
					Operator:  "invalid",
					Threshold: 0.01,
				},
			},
		}

		err := ValidateStrategy(strategy)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "invalid condition operator")
	})
}

func TestValidateWallet(t *testing.T) {
	t.Run("Valid Wallet", func(t *testing.T) {
		wallet := &Wallet{
			Address:  common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
			Balances: make(map[string]*big.Int),
		}

		err := ValidateWallet(wallet)
		assert.NoError(t, err)
	})

	t.Run("Nil Wallet", func(t *testing.T) {
		err := ValidateWallet(nil)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "wallet cannot be nil")
	})

	t.Run("Empty Address", func(t *testing.T) {
		wallet := &Wallet{
			Address:  common.Address{},
			Balances: make(map[string]*big.Int),
		}

		err := ValidateWallet(wallet)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "wallet address cannot be empty")
	})

	t.Run("Nil Balances", func(t *testing.T) {
		wallet := &Wallet{
			Address: common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		}

		err := ValidateWallet(wallet)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "wallet balances map cannot be nil")
	})
}

func TestValidateAgent(t *testing.T) {
	t.Run("Valid Agent", func(t *testing.T) {
		wallet := &Wallet{
			Address:  common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
			Balances: make(map[string]*big.Int),
		}

		strategy := Strategy{
			Type: StrategyArbitrage,
			Conditions: []Condition{
				{
					Metric:    "price_difference",
					Operator:  ">",
					Threshold: 0.01,
				},
			},
		}

		agent := &DeFiAgent{
			ID:       "test-agent",
			Name:     "Test Agent",
			Strategy: strategy,
			Wallet:   wallet,
		}

		err := ValidateAgent(agent)
		assert.NoError(t, err)
	})

	t.Run("Nil Agent", func(t *testing.T) {
		err := ValidateAgent(nil)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "agent cannot be nil")
	})

	t.Run("Empty Agent ID", func(t *testing.T) {
		wallet := &Wallet{
			Address:  common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
			Balances: make(map[string]*big.Int),
		}

		agent := &DeFiAgent{
			ID:     "",
			Name:   "Test Agent",
			Wallet: wallet,
		}

		err := ValidateAgent(agent)
		assert.Error(t, err)
		assert.True(t, IsValidationError(err))
		assert.Contains(t, err.Error(), "agent ID cannot be empty")
	})
}

func TestErrorTypeHelpers(t *testing.T) {
	t.Run("IsValidationError", func(t *testing.T) {
		err := NewDeFiError(ErrValidation, "test", nil)
		assert.True(t, IsValidationError(err))
		assert.False(t, IsStrategyError(err))
	})

	t.Run("IsStrategyError", func(t *testing.T) {
		err := NewDeFiError(ErrStrategy, "test", nil)
		assert.True(t, IsStrategyError(err))
		assert.False(t, IsBlockchainError(err))
	})

	t.Run("IsBlockchainError", func(t *testing.T) {
		err := NewDeFiError(ErrBlockchain, "test", nil)
		assert.True(t, IsBlockchainError(err))
		assert.False(t, IsValidationError(err))
	})

	t.Run("Non-DeFi Error", func(t *testing.T) {
		err := errors.New("regular error")
		assert.False(t, IsValidationError(err))
		assert.False(t, IsStrategyError(err))
		assert.False(t, IsBlockchainError(err))
	})
}

func TestCommonErrorConstructors(t *testing.T) {
	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("test_field", "invalid value")
		assert.Equal(t, ErrValidation, err.Type)
		assert.Equal(t, "invalid value", err.Message)
		assert.Equal(t, "test_field", err.Details["field"])
	})

	t.Run("NewStrategyError", func(t *testing.T) {
		err := NewStrategyError("strategy failed", StrategyArbitrage)
		assert.Equal(t, ErrStrategy, err.Type)
		assert.Equal(t, "strategy failed", err.Message)
		assert.Equal(t, "arbitrage", err.Details["strategy_type"])
	})

	t.Run("NewMarketDataError", func(t *testing.T) {
		err := NewMarketDataError("price feed unavailable", "ETH/USD")
		assert.Equal(t, ErrMarketData, err.Type)
		assert.Equal(t, "price feed unavailable", err.Message)
		assert.Equal(t, "ETH/USD", err.Details["symbol"])
	})

	t.Run("NewBlockchainError", func(t *testing.T) {
		err := NewBlockchainError("transaction failed", "swap")
		assert.Equal(t, ErrBlockchain, err.Type)
		assert.Equal(t, "transaction failed", err.Message)
		assert.Equal(t, "swap", err.Details["operation"])
	})
}
