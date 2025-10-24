# Aegis-Defi-Agent API Documentation

## Overview

The Aegis-Defi-Agent API provides RESTful endpoints for portfolio management, market data access, DeFi strategy execution, and AI agent coordination. The API is built with OpenAPI 3.0 specification and supports JSON request/response formats.

## Quick Start

### Starting the API Server

```bash
# Build and run the API server
make build-api
bin/aegis-api

# Or run directly
make run-api
```

### API Server Endpoints

- **API Documentation**: http://localhost:8080/api/docs
- **OpenAPI Spec**: http://localhost:8080/api/openapi.yaml
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:8080/metrics

## API Client Examples

### Python Client Example

```python
import requests
import json

class AegisAPIClient:
    def __init__(self, base_url="http://localhost:8080", api_key=None):
        self.base_url = base_url
        self.headers = {
            "Content-Type": "application/json",
            "Accept": "application/json"
        }
        if api_key:
            self.headers["Authorization"] = f"Bearer {api_key}"
    
    def get_portfolios(self):
        """Get all portfolios"""
        response = requests.get(f"{self.base_url}/api/v1/portfolio", headers=self.headers)
        return response.json()
    
    def create_portfolio(self, name, description=""):
        """Create a new portfolio"""
        data = {
            "name": name,
            "description": description
        }
        response = requests.post(
            f"{self.base_url}/api/v1/portfolio", 
            headers=self.headers, 
            json=data
        )
        return response.json()
    
    def get_market_data(self, symbol):
        """Get market data for a symbol"""
        response = requests.get(
            f"{self.base_url}/api/v1/market/data/{symbol}", 
            headers=self.headers
        )
        return response.json()
    
    def execute_strategy(self, strategy_id, parameters):
        """Execute a DeFi strategy"""
        data = {
            "strategy_id": strategy_id,
            "parameters": parameters
        }
        response = requests.post(
            f"{self.base_url}/api/v1/defi/strategies/execute", 
            headers=self.headers, 
            json=data
        )
        return response.json()

# Usage example
client = AegisAPIClient()

# Get all portfolios
portfolios = client.get_portfolios()
print("Portfolios:", portfolios)

# Create a new portfolio
new_portfolio = client.create_portfolio("My DeFi Portfolio", "Automated yield farming")
print("Created portfolio:", new_portfolio)

# Get market data
eth_data = client.get_market_data("ETH")
print("ETH Market Data:", eth_data)
```

### JavaScript/Node.js Client Example

```javascript
const axios = require('axios');

class AegisAPIClient {
    constructor(baseUrl = 'http://localhost:8080', apiKey = null) {
        this.baseUrl = baseUrl;
        this.client = axios.create({
            baseURL: baseUrl,
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            }
        });
        
        if (apiKey) {
            this.client.defaults.headers.common['Authorization'] = `Bearer ${apiKey}`;
        }
    }

    async getPortfolios() {
        const response = await this.client.get('/api/v1/portfolio');
        return response.data;
    }

    async createPortfolio(name, description = '') {
        const response = await this.client.post('/api/v1/portfolio', {
            name,
            description
        });
        return response.data;
    }

    async getMarketData(symbol) {
        const response = await this.client.get(`/api/v1/market/data/${symbol}`);
        return response.data;
    }

    async executeStrategy(strategyId, parameters) {
        const response = await this.client.post('/api/v1/defi/strategies/execute', {
            strategy_id: strategyId,
            parameters
        });
        return response.data;
    }
}

// Usage example
async function main() {
    const client = new AegisAPIClient();
    
    try {
        // Get all portfolios
        const portfolios = await client.getPortfolios();
        console.log('Portfolios:', portfolios);
        
        // Create a new portfolio
        const newPortfolio = await client.createPortfolio('Crypto Portfolio', 'Automated trading');
        console.log('Created portfolio:', newPortfolio);
        
        // Get market data
        const btcData = await client.getMarketData('BTC');
        console.log('BTC Market Data:', btcData);
        
    } catch (error) {
        console.error('API Error:', error.response?.data || error.message);
    }
}

main();
```

### Go Client Example

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type AegisAPIClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewAegisAPIClient(baseURL, apiKey string) *AegisAPIClient {
	return &AegisAPIClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AegisAPIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (c *AegisAPIClient) GetPortfolios() ([]map[string]interface{}, error) {
	body, err := c.doRequest("GET", "/api/v1/portfolio", nil)
	if err != nil {
		return nil, err
	}

	var portfolios []map[string]interface{}
	err = json.Unmarshal(body, &portfolios)
	return portfolios, err
}

func (c *AegisAPIClient) CreatePortfolio(name, description string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"name":        name,
		"description": description,
	}

	body, err := c.doRequest("POST", "/api/v1/portfolio", data)
	if err != nil {
		return nil, err
	}

	var portfolio map[string]interface{}
	err = json.Unmarshal(body, &portfolio)
	return portfolio, err
}

func (c *AegisAPIClient) GetMarketData(symbol string) (map[string]interface{}, error) {
	body, err := c.doRequest("GET", "/api/v1/market/data/"+symbol, nil)
	if err != nil {
		return nil, err
	}

	var marketData map[string]interface{}
	err = json.Unmarshal(body, &marketData)
	return marketData, err
}

func main() {
	client := NewAegisAPIClient("http://localhost:8080", "")

	// Get all portfolios
	portfolios, err := client.GetPortfolios()
	if err != nil {
		fmt.Printf("Error getting portfolios: %v\n", err)
		return
	}
	fmt.Printf("Portfolios: %+v\n", portfolios)

	// Create a new portfolio
	newPortfolio, err := client.CreatePortfolio("Go Portfolio", "Created via Go client")
	if err != nil {
		fmt.Printf("Error creating portfolio: %v\n", err)
		return
	}
	fmt.Printf("Created portfolio: %+v\n", newPortfolio)

	// Get market data
	ethData, err := client.GetMarketData("ETH")
	if err != nil {
		fmt.Printf("Error getting market data: %v\n", err)
		return
	}
	fmt.Printf("ETH Market Data: %+v\n", ethData)
}
```

## API Endpoints

### Portfolio Management

#### GET /api/v1/portfolio
Get all portfolios

**Response:**
```json
[
  {
    "id": "portfolio-001",
    "name": "Main Portfolio",
    "description": "Primary trading portfolio",
    "total_value": 15000.50,
    "assets": [
      {
        "symbol": "ETH",
        "amount": 2.5,
        "value": 7500.25
      }
    ],
    "created_at": "2025-01-25T10:30:00Z",
    "updated_at": "2025-01-25T10:30:00Z"
  }
]
```

#### POST /api/v1/portfolio
Create a new portfolio

**Request:**
```json
{
  "name": "My Portfolio",
  "description": "Automated yield farming"
}
```

#### GET /api/v1/portfolio/{id}
Get portfolio by ID

#### PUT /api/v1/portfolio/{id}
Update portfolio

#### DELETE /api/v1/portfolio/{id}
Delete portfolio

### Market Data

#### GET /api/v1/market/data/{symbol}
Get market data for a symbol

**Response:**
```json
{
  "symbol": "ETH",
  "price": 3000.50,
  "change_24h": 2.5,
  "volume_24h": 15000000,
  "market_cap": 360000000000,
  "last_updated": "2025-01-25T10:30:00Z"
}
```

#### GET /api/v1/market/prices
Get prices for multiple symbols

**Request:**
```json
{
  "symbols": ["BTC", "ETH", "SOL"]
}
```

### DeFi Strategies

#### GET /api/v1/defi/strategies
Get available strategies

#### POST /api/v1/defi/strategies/execute
Execute a strategy

**Request:**
```json
{
  "strategy_id": "arbitrage_eth_usdc",
  "parameters": {
    "amount": 1000,
    "min_profit": 0.01
  }
}
```

### AI Agents

#### GET /api/v1/agents
Get all agents

#### POST /api/v1/agents
Create a new agent

#### POST /api/v1/agents/{id}/tasks
Execute agent task

## Authentication

The API supports two authentication methods:

### API Key Authentication
```
Authorization: Bearer your-api-key-here
```

### JWT Token Authentication
```
Authorization: Bearer your-jwt-token-here
```

## Error Handling

All API endpoints return standard HTTP status codes:

- `200` - Success
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `500` - Internal Server Error

Error responses include detailed error messages:

```json
{
  "error": "Invalid portfolio ID",
  "code": "PORTFOLIO_NOT_FOUND",
  "details": "The specified portfolio ID does not exist"
}
```

## Rate Limiting

The API implements rate limiting to ensure fair usage:

- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per API key

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1643123456
```

## WebSocket Support

Real-time market data is available via WebSocket:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws/market');

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Market update:', data);
};

// Subscribe to symbols
ws.send(JSON.stringify({
    action: 'subscribe',
    symbols: ['BTC', 'ETH', 'SOL']
}));
```

## Testing

### Using curl

```bash
# Get all portfolios
curl -X GET http://localhost:8080/api/v1/portfolio

# Create a portfolio
curl -X POST http://localhost:8080/api/v1/portfolio \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Portfolio", "description": "API test"}'

# Get market data
curl -X GET http://localhost:8080/api/v1/market/data/BTC
```

### Health Check

```bash
curl -X GET http://localhost:8080/health
```

## Configuration

The API server can be configured via environment variables:

```bash
# Server configuration
export AEGIS_API_PORT=8080
export AEGIS_API_HOST=0.0.0.0

# Database configuration
export AEGIS_DB_HOST=localhost
export AEGIS_DB_PORT=5432

# Authentication
export AEGIS_API_KEY_SECRET=your-secret-key
```

Or via configuration file:

```yaml
server:
  port: 8080
  host: 0.0.0.0

database:
  host: localhost
  port: 5432
  name: aegis_db

authentication:
  api_key_secret: your-secret-key
```

## Monitoring

The API provides Prometheus metrics at `/metrics`:

- Request count and duration
- Error rates
- Active connections
- Portfolio metrics
- Strategy execution metrics

## Troubleshooting

### Common Issues

1. **Connection refused**: Ensure API server is running on the correct port
2. **Authentication errors**: Verify API key or JWT token
3. **Rate limiting**: Check rate limit headers and adjust request frequency
4. **Invalid requests**: Validate request body against OpenAPI specification

### Debug Mode

Enable debug logging for detailed API operations:

```bash
export AEGIS_LOG_LEVEL=debug
bin/aegis-api
```

## Support

For API-related issues and questions:

1. Check the OpenAPI documentation at `/api/docs`
2. Review server logs for detailed error information
3. Test with the provided client examples
4. Contact the development team for assistance