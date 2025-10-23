package mcpclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type BlockscoutClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

type BlockscoutResponse struct {
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
	Status  string          `json:"status"`
}

type BlockscoutTransactionResponse struct {
	Items []Transaction `json:"items"`
}

type Transaction struct {
	Hash string `json:"hash"`
	From struct {
		Hash string `json:"hash"`
	} `json:"from"`
	To struct {
		Hash string `json:"hash"`
	} `json:"to"`
	Value       string `json:"value"`
	GasUsed     string `json:"gas_used"`
	Timestamp   string `json:"timestamp"`
	BlockNumber int    `json:"block_number"`
	Method      string `json:"method"`
}

type TokenBalance struct {
	Token struct {
		Symbol   string `json:"symbol"`
		Decimals string `json:"decimals"`
	} `json:"token"`
	Value string `json:"value"`
}

func NewBlockscoutClient() *BlockscoutClient {
	return &BlockscoutClient{
		BaseURL: "https://eth-sepolia.blockscout.com/api/v2",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *BlockscoutClient) GetAddressTransactions(address string) ([]Transaction, error) {
	url := fmt.Sprintf("%s/addresses/%s/transactions", c.BaseURL, address)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}
	defer resp.Body.Close()

	var result BlockscoutTransactionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result.Items, nil
}

func (c *BlockscoutClient) GetTokenBalances(address string) ([]TokenBalance, error) {
	url := fmt.Sprintf("%s/addresses/%s/token-balances", c.BaseURL, address)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token balances: %v", err)
	}
	defer resp.Body.Close()

	var balances []TokenBalance
	if err := json.NewDecoder(resp.Body).Decode(&balances); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return balances, nil
}

func (c *BlockscoutClient) GetAddressInfo(address string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/addresses/%s", c.BaseURL, address)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch address info: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result, nil
}
