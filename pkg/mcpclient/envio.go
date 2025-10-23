package mcpclient

import (
	"net/http"
	"time"
)

type EnvioClient struct {
	baseURL    string
	httpClient *http.Client
}

type EnvioBlockData struct {
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
	Hash        string `json:"hash"`
	ParentHash  string `json:"parent_hash"`
	GasUsed     uint64 `json:"gas_used"`
	GasLimit    uint64 `json:"gas_limit"`
}

type EnvioTransaction struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"`
	GasUsed     uint64 `json:"gas_used"`
	GasPrice    string `json:"gas_price"`
	BlockNumber uint64 `json:"block_number"`
	Timestamp   int64  `json:"timestamp"`
}

type EnvioResponse struct {
	Data []map[string]interface{} `json:"data"`
}

func NewEnvioClient() *EnvioClient {
	return &EnvioClient{
		baseURL: "https://eth-sepolia.hypersync.xyz",
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (e *EnvioClient) GetLatestBlock() (*EnvioBlockData, error) {
	// For now, return mock data as Envio HyperSync requires proper GraphQL setup
	return e.GetMockBlock(), nil
}

func (e *EnvioClient) GetTransactionsByAddress(address string, limit int) ([]EnvioTransaction, error) {
	// For now, return empty array as Envio HyperSync requires proper GraphQL setup
	return []EnvioTransaction{}, nil
}

func (e *EnvioClient) GetBlockByNumber(blockNumber uint64) (*EnvioBlockData, error) {
	// For now, return mock data
	block := e.GetMockBlock()
	block.BlockNumber = blockNumber
	return block, nil
}

func (e *EnvioClient) GetLatestBlocks(limit int) ([]EnvioBlockData, error) {
	// For now, return mock data
	blocks := make([]EnvioBlockData, limit)
	for i := 0; i < limit; i++ {
		block := e.GetMockBlock()
		block.BlockNumber = block.BlockNumber - uint64(i)
		block.Timestamp = time.Now().Unix() - int64(i*15)
		blocks[i] = *block
	}
	return blocks, nil
}

func (e *EnvioClient) GetMockBlock() *EnvioBlockData {
	// Mock block data as fallback
	return &EnvioBlockData{
		BlockNumber: 5000000,
		Timestamp:   time.Now().Unix(),
		Hash:        "0x1234567890abcdef",
		ParentHash:  "0xfedcba0987654321",
		GasUsed:     15000000,
		GasLimit:    30000000,
	}
}
