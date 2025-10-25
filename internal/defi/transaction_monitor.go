package defi

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionConfirmed TransactionStatus = "confirmed"
	TransactionFailed    TransactionStatus = "failed"
	TransactionReverted  TransactionStatus = "reverted"
)

// TransactionInfo contains detailed information about a transaction
type TransactionInfo struct {
	Hash        common.Hash
	Status      TransactionStatus
	BlockNumber *big.Int
	GasUsed     uint64
	GasPrice    *big.Int
	From        common.Address
	To          *common.Address
	Value       *big.Int
	Timestamp   time.Time
	Error       string
}

// TransactionMonitor monitors blockchain transactions
type TransactionMonitor struct {
	Client       *ethclient.Client
	mu           sync.RWMutex
	transactions map[common.Hash]*TransactionInfo
	callbacks    map[common.Hash][]func(*TransactionInfo)
}

// NewTransactionMonitor creates a new transaction monitor
func NewTransactionMonitor(client *ethclient.Client) *TransactionMonitor {
	return &TransactionMonitor{
		Client:       client,
		transactions: make(map[common.Hash]*TransactionInfo),
		callbacks:    make(map[common.Hash][]func(*TransactionInfo)),
	}
}

// MonitorTransaction starts monitoring a transaction
func (tm *TransactionMonitor) MonitorTransaction(txHash common.Hash, from common.Address) (*TransactionInfo, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if already monitoring
	if info, exists := tm.transactions[txHash]; exists {
		return info, nil
	}

	// Create new transaction info
	info := &TransactionInfo{
		Hash:      txHash,
		Status:    TransactionPending,
		From:      from,
		Timestamp: time.Now(),
	}

	tm.transactions[txHash] = info

	// Start monitoring in background
	go tm.monitorTransaction(txHash)

	log.Printf("Started monitoring transaction: %s", txHash.Hex())
	return info, nil
}

// monitorTransaction continuously monitors a transaction
func (tm *TransactionMonitor) monitorTransaction(txHash common.Hash) {
	ctx := context.Background()
	maxAttempts := 60 // 5 minutes with 5-second intervals

	for attempt := 0; attempt < maxAttempts; attempt++ {
		time.Sleep(5 * time.Second)

		receipt, err := tm.Client.TransactionReceipt(ctx, txHash)
		if err != nil {
			if err.Error() == "not found" {
				// Transaction still pending
				continue
			}
			tm.updateTransactionStatus(txHash, TransactionFailed, err.Error())
			return
		}

		// Transaction found
		if receipt.Status == types.ReceiptStatusSuccessful {
			tm.updateTransactionStatus(txHash, TransactionConfirmed, "")

			// Get transaction details
			tx, _, err := tm.Client.TransactionByHash(ctx, txHash)
			if err == nil {
				tm.updateTransactionDetails(txHash, receipt, tx)
			}

			log.Printf("Transaction %s confirmed in block %d", txHash.Hex(), receipt.BlockNumber.Uint64())
			return
		} else {
			tm.updateTransactionStatus(txHash, TransactionReverted, "transaction reverted")
			log.Printf("Transaction %s reverted in block %d", txHash.Hex(), receipt.BlockNumber.Uint64())
			return
		}
	}

	// Timeout
	tm.updateTransactionStatus(txHash, TransactionFailed, "monitoring timeout")
	log.Printf("Transaction %s monitoring timeout", txHash.Hex())
}

// updateTransactionStatus updates the status of a transaction
func (tm *TransactionMonitor) updateTransactionStatus(txHash common.Hash, status TransactionStatus, errorMsg string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	info, exists := tm.transactions[txHash]
	if !exists {
		return
	}

	info.Status = status
	info.Error = errorMsg

	// Execute callbacks
	if callbacks, exists := tm.callbacks[txHash]; exists {
		for _, callback := range callbacks {
			go callback(info)
		}
	}
}

// updateTransactionDetails updates transaction details from receipt
func (tm *TransactionMonitor) updateTransactionDetails(txHash common.Hash, receipt *types.Receipt, tx *types.Transaction) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	info, exists := tm.transactions[txHash]
	if !exists {
		return
	}

	info.BlockNumber = receipt.BlockNumber
	info.GasUsed = receipt.GasUsed
	info.GasPrice = tx.GasPrice()
	info.To = tx.To()
	info.Value = tx.Value()
}

// AddCallback adds a callback for transaction status changes
func (tm *TransactionMonitor) AddCallback(txHash common.Hash, callback func(*TransactionInfo)) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.callbacks[txHash] = append(tm.callbacks[txHash], callback)
}

// GetTransactionInfo returns current transaction info
func (tm *TransactionMonitor) GetTransactionInfo(txHash common.Hash) (*TransactionInfo, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	info, exists := tm.transactions[txHash]
	return info, exists
}

// GetAllTransactions returns all monitored transactions
func (tm *TransactionMonitor) GetAllTransactions() []*TransactionInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transactions := make([]*TransactionInfo, 0, len(tm.transactions))
	for _, info := range tm.transactions {
		transactions = append(transactions, info)
	}
	return transactions
}

// GetPendingTransactions returns pending transactions
func (tm *TransactionMonitor) GetPendingTransactions() []*TransactionInfo {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var pending []*TransactionInfo
	for _, info := range tm.transactions {
		if info.Status == TransactionPending {
			pending = append(pending, info)
		}
	}
	return pending
}

// GetTransactionCost calculates the total cost of a transaction
func (tm *TransactionMonitor) GetTransactionCost(txHash common.Hash) (*big.Int, error) {
	info, exists := tm.GetTransactionInfo(txHash)
	if !exists {
		return nil, fmt.Errorf("transaction not found")
	}

	if info.GasUsed == 0 || info.GasPrice == nil {
		return nil, fmt.Errorf("transaction not confirmed yet")
	}

	gasUsed := big.NewInt(int64(info.GasUsed))
	totalCost := new(big.Int).Mul(gasUsed, info.GasPrice)

	return totalCost, nil
}

// WaitForConfirmation waits for transaction confirmation
func (tm *TransactionMonitor) WaitForConfirmation(txHash common.Hash, timeout time.Duration) (*TransactionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction confirmation")
		case <-ticker.C:
			info, exists := tm.GetTransactionInfo(txHash)
			if !exists {
				continue
			}

			if info.Status == TransactionConfirmed {
				return info, nil
			} else if info.Status == TransactionFailed || info.Status == TransactionReverted {
				return nil, fmt.Errorf("transaction failed: %s", info.Error)
			}
		}
	}
}

// CleanupCompleted removes completed transactions from monitoring
func (tm *TransactionMonitor) CleanupCompleted() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	count := 0
	for hash, info := range tm.transactions {
		if info.Status != TransactionPending {
			delete(tm.transactions, hash)
			delete(tm.callbacks, hash)
			count++
		}
	}

	log.Printf("Cleaned up %d completed transactions", count)
	return count
}

// GetTransactionStats returns statistics about monitored transactions
func (tm *TransactionMonitor) GetTransactionStats() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	pending := 0
	confirmed := 0
	failed := 0
	reverted := 0

	for _, info := range tm.transactions {
		switch info.Status {
		case TransactionPending:
			pending++
		case TransactionConfirmed:
			confirmed++
		case TransactionFailed:
			failed++
		case TransactionReverted:
			reverted++
		}
	}

	stats := map[string]interface{}{
		"total":     len(tm.transactions),
		"pending":   pending,
		"confirmed": confirmed,
		"failed":    failed,
		"reverted":  reverted,
	}

	return stats
}
