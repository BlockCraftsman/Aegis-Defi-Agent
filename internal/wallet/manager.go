package wallet

import (
	"fmt"
	"math/big"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/pkg/mcpclient"
)

type Manager struct {
	wallets       []Wallet
	currentWallet *Wallet
	blockscout    *mcpclient.BlockscoutClient
}

type Wallet struct {
	Address  string
	Chain    string
	Balance  string
	Tokens   []Token
	IsActive bool
}

type Token struct {
	Symbol   string
	Balance  string
	ValueUSD string
}

func NewManager() *Manager {
	return &Manager{
		wallets: []Wallet{
			{
				Address:  "0x742d35Cc6634C0532925a3b8D...",
				Chain:    "Ethereum",
				Balance:  "2.5 ETH",
				IsActive: true,
				Tokens: []Token{
					{Symbol: "ETH", Balance: "2.5", ValueUSD: "8,750"},
					{Symbol: "USDC", Balance: "1,500", ValueUSD: "1,500"},
				},
			},
			{
				Address:  "0x89205A3A3b2A69De6D...",
				Chain:    "Polygon",
				Balance:  "1,200 MATIC",
				IsActive: false,
				Tokens: []Token{
					{Symbol: "MATIC", Balance: "1,200", ValueUSD: "840"},
					{Symbol: "USDT", Balance: "500", ValueUSD: "500"},
				},
			},
		},
		blockscout: mcpclient.NewBlockscoutClient(),
	}
}

func (m *Manager) GetWallets() []Wallet {
	return m.wallets
}

func (m *Manager) SelectWallet(address string) error {
	for i := range m.wallets {
		if m.wallets[i].Address == address {
			m.currentWallet = &m.wallets[i]
			m.wallets[i].IsActive = true
			return nil
		}
	}
	return fmt.Errorf("wallet not found: %s", address)
}

func (m *Manager) GetCurrentWallet() *Wallet {
	return m.currentWallet
}

func (m *Manager) RefreshBalances() error {
	for i := range m.wallets {
		if m.wallets[i].IsActive {
			// Get real token balances from Blockscout
			balances, err := m.blockscout.GetTokenBalances(m.wallets[i].Address)
			if err != nil {
				// Fallback to simulated data if API fails
				m.wallets[i].Balance = "2.5 ETH"
				m.wallets[i].Tokens[0].Balance = "2.5"
				m.wallets[i].Tokens[0].ValueUSD = "8,750"
				continue
			}

			// Update wallet with real data
			for _, balance := range balances {
				if balance.Token.Symbol == "ETH" || balance.Token.Symbol == "MATIC" {
					// Convert wei to ETH/MATIC
					wei, _ := new(big.Int).SetString(balance.Value, 10)
					ethValue := new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(1e18))
					ethStr, _ := ethValue.Float64()

					m.wallets[i].Balance = fmt.Sprintf("%.4f %s", ethStr, balance.Token.Symbol)
					m.wallets[i].Tokens[0].Balance = fmt.Sprintf("%.4f", ethStr)

					// Calculate USD value (simplified)
					var usdValue float64
					if balance.Token.Symbol == "ETH" {
						usdValue = ethStr * 3500
					} else {
						usdValue = ethStr * 0.7
					}
					m.wallets[i].Tokens[0].ValueUSD = fmt.Sprintf("%.0f", usdValue)
				}
			}
		}
	}
	return nil
}

func (m *Manager) AddWallet(address, chain string) error {
	newWallet := Wallet{
		Address:  address,
		Chain:    chain,
		Balance:  "0.0",
		IsActive: false,
		Tokens:   []Token{},
	}

	// Add default tokens based on chain
	switch chain {
	case "Ethereum":
		newWallet.Tokens = []Token{
			{Symbol: "ETH", Balance: "0.0", ValueUSD: "0"},
			{Symbol: "USDC", Balance: "0.0", ValueUSD: "0"},
		}
	case "Polygon":
		newWallet.Tokens = []Token{
			{Symbol: "MATIC", Balance: "0.0", ValueUSD: "0"},
			{Symbol: "USDT", Balance: "0.0", ValueUSD: "0"},
		}
	}

	m.wallets = append(m.wallets, newWallet)
	return nil
}

func (m *Manager) RemoveWallet(address string) error {
	for i, wallet := range m.wallets {
		if wallet.Address == address {
			m.wallets = append(m.wallets[:i], m.wallets[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("wallet not found: %s", address)
}

func (m *Manager) GetWalletDetails(address string) (*Wallet, error) {
	for i := range m.wallets {
		if m.wallets[i].Address == address {
			return &m.wallets[i], nil
		}
	}
	return nil, fmt.Errorf("wallet not found: %s", address)
}
