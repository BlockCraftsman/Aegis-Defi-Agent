package defi

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

// Error types for DeFi operations
type ErrorType string

const (
	ErrValidation     ErrorType = "VALIDATION_ERROR"
	ErrStrategy       ErrorType = "STRATEGY_ERROR"
	ErrMarketData     ErrorType = "MARKET_DATA_ERROR"
	ErrBlockchain     ErrorType = "BLOCKCHAIN_ERROR"
	ErrRiskManagement ErrorType = "RISK_MANAGEMENT_ERROR"
	ErrWallet         ErrorType = "WALLET_ERROR"
	ErrContract       ErrorType = "CONTRACT_ERROR"
	ErrNetwork        ErrorType = "NETWORK_ERROR"
	ErrConfiguration  ErrorType = "CONFIGURATION_ERROR"
	ErrExecution      ErrorType = "EXECUTION_ERROR"
)

// DeFiError represents a structured error in DeFi operations
type DeFiError struct {
	Type    ErrorType
	Message string
	Details map[string]interface{}
	Cause   error
}

func (e *DeFiError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (cause: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *DeFiError) Unwrap() error {
	return e.Cause
}

// NewDeFiError creates a new DeFi error
func NewDeFiError(errorType ErrorType, message string, details map[string]interface{}) *DeFiError {
	return &DeFiError{
		Type:    errorType,
		Message: message,
		Details: details,
	}
}

// WrapDeFiError wraps an existing error with DeFi error context
func WrapDeFiError(errorType ErrorType, message string, cause error, details map[string]interface{}) *DeFiError {
	return &DeFiError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Details: details,
	}
}

// Validation functions
func ValidateStrategy(strategy Strategy) error {
	if strategy.Type == "" {
		return NewDeFiError(ErrValidation, "strategy type cannot be empty", map[string]interface{}{
			"field": "Type",
		})
	}

	if len(strategy.Conditions) == 0 && strategy.IsEnabled {
		return NewDeFiError(ErrValidation, "enabled strategy must have conditions", map[string]interface{}{
			"strategy_type": string(strategy.Type),
		})
	}

	for i, condition := range strategy.Conditions {
		if condition.Metric == "" {
			return NewDeFiError(ErrValidation, "condition metric cannot be empty", map[string]interface{}{
				"condition_index": i,
			})
		}

		if condition.Operator == "" {
			return NewDeFiError(ErrValidation, "condition operator cannot be empty", map[string]interface{}{
				"condition_index": i,
				"metric":          condition.Metric,
			})
		}

		if !isValidOperator(condition.Operator) {
			return NewDeFiError(ErrValidation, "invalid condition operator", map[string]interface{}{
				"condition_index": i,
				"metric":          condition.Metric,
				"operator":        condition.Operator,
				"valid_operators": []string{">", "<", "==", ">=", "<="},
			})
		}
	}

	return nil
}

func ValidateWallet(wallet *Wallet) error {
	if wallet == nil {
		return NewDeFiError(ErrValidation, "wallet cannot be nil", nil)
	}

	if wallet.Address == (common.Address{}) {
		return NewDeFiError(ErrValidation, "wallet address cannot be empty", nil)
	}

	if wallet.Balances == nil {
		return NewDeFiError(ErrValidation, "wallet balances map cannot be nil", nil)
	}

	return nil
}

func ValidateAgent(agent *DeFiAgent) error {
	if agent == nil {
		return NewDeFiError(ErrValidation, "agent cannot be nil", nil)
	}

	if agent.ID == "" {
		return NewDeFiError(ErrValidation, "agent ID cannot be empty", nil)
	}

	if agent.Name == "" {
		return NewDeFiError(ErrValidation, "agent name cannot be empty", nil)
	}

	if err := ValidateStrategy(agent.Strategy); err != nil {
		return err
	}

	if err := ValidateWallet(agent.Wallet); err != nil {
		return err
	}

	return nil
}

// Helper functions
func isValidOperator(operator string) bool {
	validOperators := []string{">", "<", "==", ">=", "<="}
	for _, validOp := range validOperators {
		if operator == validOp {
			return true
		}
	}
	return false
}

// Common error constructors
func NewValidationError(field, message string) *DeFiError {
	return NewDeFiError(ErrValidation, message, map[string]interface{}{
		"field": field,
	})
}

func NewStrategyError(message string, strategyType StrategyType) *DeFiError {
	return NewDeFiError(ErrStrategy, message, map[string]interface{}{
		"strategy_type": string(strategyType),
	})
}

func NewMarketDataError(message string, symbol string) *DeFiError {
	return NewDeFiError(ErrMarketData, message, map[string]interface{}{
		"symbol": symbol,
	})
}

func NewBlockchainError(message string, operation string) *DeFiError {
	return NewDeFiError(ErrBlockchain, message, map[string]interface{}{
		"operation": operation,
	})
}

// Error checking helpers
func IsValidationError(err error) bool {
	var defiErr *DeFiError
	if errors.As(err, &defiErr) {
		return defiErr.Type == ErrValidation
	}
	return false
}

func IsStrategyError(err error) bool {
	var defiErr *DeFiError
	if errors.As(err, &defiErr) {
		return defiErr.Type == ErrStrategy
	}
	return false
}

func IsBlockchainError(err error) bool {
	var defiErr *DeFiError
	if errors.As(err, &defiErr) {
		return defiErr.Type == ErrBlockchain
	}
	return false
}
