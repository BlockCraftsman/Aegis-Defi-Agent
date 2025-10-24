package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	assert.NotNil(t, manager)
	assert.Len(t, manager.wallets, 2)
	assert.NotNil(t, manager.blockscout)

	// Check default wallets
	wallets := manager.GetWallets()
	assert.Len(t, wallets, 2)

	// Check Ethereum wallet
	ethWallet := wallets[0]
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D...", ethWallet.Address)
	assert.Equal(t, "Ethereum", ethWallet.Chain)
	assert.Equal(t, "2.5 ETH", ethWallet.Balance)
	assert.True(t, ethWallet.IsActive)
	assert.Len(t, ethWallet.Tokens, 2)

	// Check Polygon wallet
	polygonWallet := wallets[1]
	assert.Equal(t, "0x89205A3A3b2A69De6D...", polygonWallet.Address)
	assert.Equal(t, "Polygon", polygonWallet.Chain)
	assert.Equal(t, "1,200 MATIC", polygonWallet.Balance)
	assert.False(t, polygonWallet.IsActive)
	assert.Len(t, polygonWallet.Tokens, 2)
}

func TestSelectWallet(t *testing.T) {
	manager := NewManager()

	// Test selecting existing wallet
	err := manager.SelectWallet("0x742d35Cc6634C0532925a3b8D...")
	require.NoError(t, err)

	currentWallet := manager.GetCurrentWallet()
	assert.NotNil(t, currentWallet)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D...", currentWallet.Address)
	assert.True(t, currentWallet.IsActive)

	// Test selecting non-existent wallet
	err = manager.SelectWallet("non-existent-address")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wallet not found")
}

func TestAddWallet(t *testing.T) {
	manager := NewManager()

	// Test adding Ethereum wallet
	err := manager.AddWallet("0x1234567890abcdef", "Ethereum")
	require.NoError(t, err)

	wallets := manager.GetWallets()
	assert.Len(t, wallets, 3)

	// Check the new wallet
	newWallet, err := manager.GetWalletDetails("0x1234567890abcdef")
	require.NoError(t, err)
	assert.Equal(t, "0x1234567890abcdef", newWallet.Address)
	assert.Equal(t, "Ethereum", newWallet.Chain)
	assert.Equal(t, "0.0", newWallet.Balance)
	assert.False(t, newWallet.IsActive)
	assert.Len(t, newWallet.Tokens, 2)

	// Test adding Polygon wallet
	err = manager.AddWallet("0xfedcba0987654321", "Polygon")
	require.NoError(t, err)

	wallets = manager.GetWallets()
	assert.Len(t, wallets, 4)

	polygonWallet, err := manager.GetWalletDetails("0xfedcba0987654321")
	require.NoError(t, err)
	assert.Equal(t, "Polygon", polygonWallet.Chain)
	assert.Len(t, polygonWallet.Tokens, 2)
}

func TestRemoveWallet(t *testing.T) {
	manager := NewManager()

	// Test removing existing wallet
	err := manager.RemoveWallet("0x742d35Cc6634C0532925a3b8D...")
	require.NoError(t, err)

	wallets := manager.GetWallets()
	assert.Len(t, wallets, 1)

	// Test removing non-existent wallet
	err = manager.RemoveWallet("non-existent-address")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wallet not found")
}

func TestGetWalletDetails(t *testing.T) {
	manager := NewManager()

	// Test getting existing wallet
	wallet, err := manager.GetWalletDetails("0x742d35Cc6634C0532925a3b8D...")
	require.NoError(t, err)
	assert.Equal(t, "0x742d35Cc6634C0532925a3b8D...", wallet.Address)
	assert.Equal(t, "Ethereum", wallet.Chain)

	// Test getting non-existent wallet
	wallet, err = manager.GetWalletDetails("non-existent-address")
	require.Error(t, err)
	assert.Nil(t, wallet)
	assert.Contains(t, err.Error(), "wallet not found")
}

func TestRefreshBalances(t *testing.T) {
	manager := NewManager()

	// Select a wallet first
	err := manager.SelectWallet("0x742d35Cc6634C0532925a3b8D...")
	require.NoError(t, err)

	// Test refreshing balances
	err = manager.RefreshBalances()
	require.NoError(t, err)

	// Check that balances were updated (even if simulated)
	currentWallet := manager.GetCurrentWallet()
	assert.NotNil(t, currentWallet)
	assert.Equal(t, "2.5 ETH", currentWallet.Balance)
}

func TestTokenStructure(t *testing.T) {
	token := Token{
		Symbol:   "ETH",
		Balance:  "2.5",
		ValueUSD: "8,750",
	}

	assert.Equal(t, "ETH", token.Symbol)
	assert.Equal(t, "2.5", token.Balance)
	assert.Equal(t, "8,750", token.ValueUSD)
}

func TestWalletStructure(t *testing.T) {
	wallet := Wallet{
		Address:  "0x1234567890abcdef",
		Chain:    "Ethereum",
		Balance:  "1.5 ETH",
		IsActive: true,
		Tokens: []Token{
			{Symbol: "ETH", Balance: "1.5", ValueUSD: "5,250"},
			{Symbol: "USDC", Balance: "1000", ValueUSD: "1000"},
		},
	}

	assert.Equal(t, "0x1234567890abcdef", wallet.Address)
	assert.Equal(t, "Ethereum", wallet.Chain)
	assert.Equal(t, "1.5 ETH", wallet.Balance)
	assert.True(t, wallet.IsActive)
	assert.Len(t, wallet.Tokens, 2)
}
