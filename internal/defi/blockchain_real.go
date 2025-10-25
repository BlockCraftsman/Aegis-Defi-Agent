package defi

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// RealBlockchainManager handles real blockchain interactions
type RealBlockchainManager struct {
	Client     *ethclient.Client
	PrivateKey string
	ChainID    *big.Int
	Address    common.Address
}

// NewRealBlockchainManager creates a new blockchain manager with real connection
func NewRealBlockchainManager(privateKey string) (*RealBlockchainManager, error) {
	rpcURL := getRealRPCURL()

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to blockchain: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %v", err)
	}

	// Get address from private key
	address, err := getAddressFromPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	log.Printf("Connected to blockchain at %s", rpcURL)
	log.Printf("Chain ID: %s", chainID.String())
	log.Printf("Wallet address: %s", address.Hex())

	return &RealBlockchainManager{
		Client:     client,
		PrivateKey: privateKey,
		ChainID:    chainID,
		Address:    address,
	}, nil
}

// getRealRPCURL returns the appropriate RPC URL
func getRealRPCURL() string {
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

// getAddressFromPrivateKey derives address from private key
func getAddressFromPrivateKey(privateKey string) (common.Address, error) {
	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return common.Address{}, fmt.Errorf("invalid private key: %v", err)
	}

	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

// GetBalance returns the ETH balance of the wallet
func (bm *RealBlockchainManager) GetBalance() (*big.Int, error) {
	balance, err := bm.Client.BalanceAt(context.Background(), bm.Address, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %v", err)
	}
	return balance, nil
}

// GetTokenBalance returns the balance of an ERC20 token
func (bm *RealBlockchainManager) GetTokenBalance(tokenAddress common.Address) (*big.Int, error) {
	// ERC20 balanceOf function
	data, err := packBalanceOfCall(bm.Address)
	if err != nil {
		return nil, err
	}

	result, err := bm.Client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %v", err)
	}

	balance := new(big.Int).SetBytes(result)
	return balance, nil
}

// SendTransaction sends a transaction
func (bm *RealBlockchainManager) SendTransaction(to common.Address, value *big.Int, data []byte) (*types.Transaction, error) {
	// Get current nonce
	nonce, err := bm.Client.PendingNonceAt(context.Background(), bm.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	// Get gas price
	gasPrice, err := bm.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %v", err)
	}

	// Estimate gas
	gasLimit, err := bm.Client.EstimateGas(context.Background(), ethereum.CallMsg{
		From:  bm.Address,
		To:    &to,
		Value: value,
		Data:  data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %v", err)
	}

	// Create transaction
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, data)

	// Sign transaction
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(bm.PrivateKey, "0x"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(bm.ChainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send transaction
	err = bm.Client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %v", err)
	}

	log.Printf("Transaction sent: %s", signedTx.Hash().Hex())
	return signedTx, nil
}

// WaitForTransaction waits for transaction confirmation
func (bm *RealBlockchainManager) WaitForTransaction(txHash common.Hash) (*types.Receipt, error) {
	ctx := context.Background()

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(ctx, bm.Client, &types.Transaction{})
	if err != nil {
		return nil, fmt.Errorf("failed to wait for transaction: %v", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		return nil, fmt.Errorf("transaction failed: %s", txHash.Hex())
	}

	log.Printf("Transaction confirmed in block %d", receipt.BlockNumber.Uint64())
	return receipt, nil
}

// packBalanceOfCall packs balanceOf function call
func packBalanceOfCall(address common.Address) ([]byte, error) {
	// balanceOf function selector + padded address
	data := make([]byte, 36)
	copy(data[0:4], crypto.Keccak256([]byte("balanceOf(address)"))[0:4])
	copy(data[4:36], address.Bytes())
	return data, nil
}

// GetGasPrice returns current gas price
func (bm *RealBlockchainManager) GetGasPrice() (*big.Int, error) {
	return bm.Client.SuggestGasPrice(context.Background())
}

// GetBlockNumber returns current block number
func (bm *RealBlockchainManager) GetBlockNumber() (uint64, error) {
	return bm.Client.BlockNumber(context.Background())
}
