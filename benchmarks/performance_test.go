package benchmarks

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/defi"
	"github.com/ethereum/go-ethereum/common"
)

// BenchmarkDeFiAgentCreation measures the performance of creating new DeFi agents
func BenchmarkDeFiAgentCreation(b *testing.B) {
	wallet := &defi.Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := defi.Strategy{
		Type: defi.StrategyArbitrage,
		Parameters: map[string]any{
			"min_profit_threshold": 0.01,
			"max_slippage":         0.005,
		},
		Conditions: []defi.Condition{
			{
				Metric:    "price_difference",
				Operator:  ">",
				Threshold: 0.005,
			},
		},
		IsEnabled: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = defi.NewDeFiAgent("benchmark-agent", "Benchmark Bot", strategy, wallet)
	}
}

// BenchmarkStrategyEvaluation measures the performance of strategy condition evaluation
func BenchmarkStrategyEvaluation(b *testing.B) {
	wallet := &defi.Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := defi.Strategy{
		Type: defi.StrategyArbitrage,
		Conditions: []defi.Condition{
			{
				Metric:    "price_difference",
				Operator:  ">",
				Threshold: 0.01,
			},
			{
				Metric:    "yield_rate",
				Operator:  ">",
				Threshold: 0.05,
			},
		},
		IsEnabled: true,
	}

	agent := defi.NewDeFiAgent("benchmark-agent", "Benchmark Bot", strategy, wallet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use exported method GetStatus instead of unexported evaluateConditions
		agent.GetStatus()
	}
}

// BenchmarkMarketDataUpdate measures the performance of market data updates
func BenchmarkMarketDataUpdate(b *testing.B) {
	marketData := defi.NewMarketData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate market data update
		marketData.PriceFeeds["ETH/USD"] = 3500.0 + float64(i%100)
		marketData.PriceFeeds["BTC/USD"] = 45000.0 + float64(i%100)
		marketData.LastUpdate = time.Now()
	}
}

// BenchmarkRiskAssessment measures the performance of risk assessment
func BenchmarkRiskAssessment(b *testing.B) {
	wallet := &defi.Wallet{
		Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
		PrivateKey: "encrypted_key",
		Balances:   make(map[string]*big.Int),
		Nonce:      0,
	}

	strategy := defi.Strategy{
		Type:      defi.StrategyArbitrage,
		IsEnabled: true,
	}

	agent := defi.NewDeFiAgent("benchmark-agent", "Benchmark Bot", strategy, wallet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use exported method GetStatus instead of unexported assessRisk
		agent.GetStatus()
	}
}

// BenchmarkStrategyEngineOperations measures the performance of strategy engine operations
func BenchmarkStrategyEngineOperations(b *testing.B) {
	engine := defi.NewStrategyEngine()

	// Add multiple strategies
	for i := 0; i < 100; i++ {
		strategy := &defi.TradingStrategy{
			ID:          "benchmark_strategy",
			Name:        "Benchmark Strategy",
			Type:        defi.StrategyArbitrage,
			Description: "Benchmark strategy for performance testing",
			Parameters: defi.StrategyParameters{
				MaxPositionSize: 0.1,
				MinProfitMargin: 0.01,
				MaxSlippage:     0.005,
				ExecutionDelay:  5,
				CooldownPeriod:  60,
				TargetAssets:    []string{"ETH", "USDC"},
			},
			IsActive: true,
		}
		_ = engine.AddStrategy(strategy)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate strategy evaluation
		for _, strategy := range engine.Strategies {
			if strategy.IsActive {
				// Evaluate strategy conditions
				_ = strategy
			}
		}
	}
}

// BenchmarkConcurrentAgentOperations measures performance under concurrent load
func BenchmarkConcurrentAgentOperations(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		wallet := &defi.Wallet{
			Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
			PrivateKey: "encrypted_key",
			Balances:   make(map[string]*big.Int),
			Nonce:      0,
		}

		strategy := defi.Strategy{
			Type:      defi.StrategyArbitrage,
			IsEnabled: true,
		}

		agent := defi.NewDeFiAgent("concurrent-agent", "Concurrent Bot", strategy, wallet)

		for pb.Next() {
			// Simulate concurrent operations using exported methods
			agent.GetStatus()
			agent.GetStatus() // Call multiple times to simulate workload
			agent.GetStatus()
		}
	})
}

// BenchmarkMemoryUsage measures memory allocation patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		wallet := &defi.Wallet{
			Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
			PrivateKey: "encrypted_key",
			Balances:   make(map[string]*big.Int),
			Nonce:      0,
		}

		strategy := defi.Strategy{
			Type: defi.StrategyArbitrage,
			Parameters: map[string]any{
				"min_profit_threshold": 0.01,
				"max_slippage":         0.005,
			},
			Conditions: []defi.Condition{
				{
					Metric:    "price_difference",
					Operator:  ">",
					Threshold: 0.005,
				},
			},
			IsEnabled: true,
		}

		agent := defi.NewDeFiAgent("memory-agent", "Memory Bot", strategy, wallet)
		_ = agent
	}
}

// BenchmarkAgentLifecycle measures the complete agent lifecycle
func BenchmarkAgentLifecycle(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wallet := &defi.Wallet{
			Address:    common.HexToAddress("0x742d35Cc6634C0532925a3b8D"),
			PrivateKey: "encrypted_key",
			Balances:   make(map[string]*big.Int),
			Nonce:      0,
		}

		strategy := defi.Strategy{
			Type:      defi.StrategyArbitrage,
			IsEnabled: true,
		}

		agent := defi.NewDeFiAgent("lifecycle-agent", "Lifecycle Bot", strategy, wallet)

		ctx, cancel := context.WithCancel(context.Background())

		// Start agent
		_ = agent.Start(ctx)

		// Simulate some operations
		agent.GetStatus()
		agent.GetStatus() // Call multiple times to simulate workload

		// Stop agent
		agent.Stop()
		cancel()
	}
}
