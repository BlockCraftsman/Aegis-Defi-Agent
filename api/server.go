package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/config"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/logging"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/monitoring"
	"github.com/BlockCraftsman/Aegis-Defi-Agent/internal/portfolio"
	"github.com/gorilla/mux"
)

// Server represents the API server
type Server struct {
	router     *mux.Router
	server     *http.Server
	logger     logging.Logger
	monitor    *monitoring.Monitor
	portfolios *portfolio.PortfolioManager
	startTime  time.Time
	mu         sync.RWMutex
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, logger logging.Logger, monitor *monitoring.Monitor, portfolios *portfolio.PortfolioManager) *Server {
	s := &Server{
		router:     mux.NewRouter(),
		logger:     logger,
		monitor:    monitor,
		portfolios: portfolios,
		startTime:  time.Now(),
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health and monitoring
	s.router.HandleFunc("/health", s.healthHandler).Methods("GET")
	s.router.HandleFunc("/metrics", s.metricsHandler).Methods("GET")

	// API documentation
	s.router.HandleFunc("/api/docs", s.docsHandler).Methods("GET")
	s.router.HandleFunc("/api/openapi.yaml", s.openapiHandler).Methods("GET")

	// API v1 routes
	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	// Portfolio endpoints
	apiV1.HandleFunc("/portfolio", s.listPortfolios).Methods("GET")
	apiV1.HandleFunc("/portfolio", s.createPortfolio).Methods("POST")
	apiV1.HandleFunc("/portfolio/{portfolioId}", s.getPortfolio).Methods("GET")
	apiV1.HandleFunc("/portfolio/{portfolioId}", s.deletePortfolio).Methods("DELETE")
	apiV1.HandleFunc("/portfolio/{portfolioId}/assets", s.getPortfolioAssets).Methods("GET")
	apiV1.HandleFunc("/portfolio/{portfolioId}/assets", s.addAssetToPortfolio).Methods("POST")
	apiV1.HandleFunc("/portfolio/{portfolioId}/rebalance", s.rebalancePortfolio).Methods("POST")
	apiV1.HandleFunc("/portfolio/{portfolioId}/risk", s.getPortfolioRisk).Methods("GET")

	// Market endpoints
	apiV1.HandleFunc("/market/data", s.getMarketData).Methods("GET")

	// DeFi endpoints
	apiV1.HandleFunc("/defi/strategies", s.listStrategies).Methods("GET")
	apiV1.HandleFunc("/defi/strategies/{strategyId}/execute", s.executeStrategy).Methods("POST")

	// Agent endpoints
	apiV1.HandleFunc("/agents", s.listAgents).Methods("GET")
	apiV1.HandleFunc("/agents/{agentId}/execute", s.executeAgent).Methods("POST")

	// Middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
}

// Start starts the API server
func (s *Server) Start(port int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: s.router,
	}

	s.logger.Info("Starting API server",
		logging.WithInt("port", port),
	)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("API server failed",
				logging.WithError(err),
			)
		}
	}()

	return nil
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Version:   "2.0.0",
		Uptime:    time.Since(s.startTime).Seconds(),
	}

	s.respondJSON(w, http.StatusOK, response)
}

// metricsHandler handles metrics requests
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would serve Prometheus metrics
	// For now, we'll return a simple message
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# Aegis DeFi Agent Metrics\n# Metrics are available via Prometheus\n"))
}

// docsHandler serves API documentation
func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	// Serve Swagger UI or redirect to OpenAPI spec
	http.Redirect(w, r, "/api/openapi.yaml", http.StatusFound)
}

// openapiHandler serves the OpenAPI specification
func (s *Server) openapiHandler(w http.ResponseWriter, r *http.Request) {
	// Read the OpenAPI spec file
	openapiPath := filepath.Join("api", "openapi.yaml")
	data, err := os.ReadFile(openapiPath)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to read OpenAPI specification")
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// Portfolio handlers
func (s *Server) listPortfolios(w http.ResponseWriter, r *http.Request) {
	portfolios := s.portfolios.ListPortfolios()

	// Convert to API response format
	var response []Portfolio
	for _, p := range portfolios {
		stats := p.GetPortfolioStats()
		response = append(response, Portfolio{
			ID:          p.ID,
			Name:        p.Name,
			TotalValue:  toFloat(stats.TotalValue),
			CashBalance: toFloat(stats.CashBalance),
			RiskProfile: RiskProfile{
				Type:              string(p.RiskProfile.Type),
				MaxDrawdown:       p.RiskProfile.MaxDrawdown,
				MaxPositionSize:   p.RiskProfile.MaxPositionSize,
				MaxLeverage:       p.RiskProfile.MaxLeverage,
				TargetAllocations: p.RiskProfile.TargetAllocations,
			},
			Assets:        []Asset{},    // Would need conversion
			Positions:     []Position{}, // Would need conversion
			LastRebalance: p.LastRebalance,
			CreatedAt:     time.Now(), // Would need to track creation time
		})
	}

	s.respondJSON(w, http.StatusOK, response)
}

func (s *Server) createPortfolio(w http.ResponseWriter, r *http.Request) {
	var req CreatePortfolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Convert to internal risk profile
	riskProfile := portfolio.RiskProfile{
		Type:              portfolio.RiskType(req.RiskProfile.Type),
		MaxDrawdown:       req.RiskProfile.MaxDrawdown,
		MaxPositionSize:   req.RiskProfile.MaxPositionSize,
		MaxLeverage:       req.RiskProfile.MaxLeverage,
		TargetAllocations: req.RiskProfile.TargetAllocations,
	}

	p, err := s.portfolios.CreatePortfolio(req.ID, req.Name, riskProfile)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Set initial cash if provided
	if req.InitialCash > 0 {
		p.CashBalance = big.NewFloat(req.InitialCash)
	}

	stats := p.GetPortfolioStats()
	response := Portfolio{
		ID:          p.ID,
		Name:        p.Name,
		TotalValue:  toFloat(stats.TotalValue),
		CashBalance: toFloat(stats.CashBalance),
		RiskProfile: RiskProfile{
			Type:              string(p.RiskProfile.Type),
			MaxDrawdown:       p.RiskProfile.MaxDrawdown,
			MaxPositionSize:   p.RiskProfile.MaxPositionSize,
			MaxLeverage:       p.RiskProfile.MaxLeverage,
			TargetAllocations: p.RiskProfile.TargetAllocations,
		},
		Assets:        []Asset{},
		Positions:     []Position{},
		LastRebalance: p.LastRebalance,
		CreatedAt:     time.Now(),
	}

	s.respondJSON(w, http.StatusCreated, response)
}

func (s *Server) getPortfolio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["portfolioId"]

	p, err := s.portfolios.GetPortfolio(portfolioID)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "Portfolio not found")
		return
	}

	stats := p.GetPortfolioStats()
	response := Portfolio{
		ID:          p.ID,
		Name:        p.Name,
		TotalValue:  toFloat(stats.TotalValue),
		CashBalance: toFloat(stats.CashBalance),
		RiskProfile: RiskProfile{
			Type:              string(p.RiskProfile.Type),
			MaxDrawdown:       p.RiskProfile.MaxDrawdown,
			MaxPositionSize:   p.RiskProfile.MaxPositionSize,
			MaxLeverage:       p.RiskProfile.MaxLeverage,
			TargetAllocations: p.RiskProfile.TargetAllocations,
		},
		Assets:        []Asset{},    // Would need conversion
		Positions:     []Position{}, // Would need conversion
		LastRebalance: p.LastRebalance,
		CreatedAt:     time.Now(), // Would need to track creation time
	}

	s.respondJSON(w, http.StatusOK, response)
}

func (s *Server) deletePortfolio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	portfolioID := vars["portfolioId"]

	if err := s.portfolios.DeletePortfolio(portfolioID); err != nil {
		s.respondError(w, http.StatusNotFound, "Portfolio not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Other handlers would be implemented similarly...

// Helper methods
func (s *Server) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response",
			logging.WithError(err),
		)
	}
}

func (s *Server) respondError(w http.ResponseWriter, status int, message string) {
	s.respondJSON(w, status, ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		s.logger.Info("API request",
			logging.WithString("method", r.Method),
			logging.WithString("path", r.URL.Path),
			logging.WithString("remote_addr", r.RemoteAddr),
			logging.WithDuration("duration", time.Since(start)),
		)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Helper function to convert big.Float to float64
func toFloat(bf *big.Float) float64 {
	if bf == nil {
		return 0
	}
	f, _ := bf.Float64()
	return f
}

// API request/response types
// These types match the OpenAPI specification

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
	Uptime    float64   `json:"uptime"`
}

type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

type Portfolio struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	TotalValue    float64     `json:"totalValue"`
	CashBalance   float64     `json:"cashBalance"`
	RiskProfile   RiskProfile `json:"riskProfile"`
	Assets        []Asset     `json:"assets"`
	Positions     []Position  `json:"positions"`
	LastRebalance time.Time   `json:"lastRebalance"`
	CreatedAt     time.Time   `json:"createdAt"`
}

type CreatePortfolioRequest struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	RiskProfile RiskProfile `json:"riskProfile"`
	InitialCash float64     `json:"initialCash,omitempty"`
}

type RiskProfile struct {
	Type              string             `json:"type"`
	MaxDrawdown       float64            `json:"maxDrawdown,omitempty"`
	MaxPositionSize   float64            `json:"maxPositionSize,omitempty"`
	MaxLeverage       float64            `json:"maxLeverage,omitempty"`
	TargetAllocations map[string]float64 `json:"targetAllocations,omitempty"`
}

type Asset struct {
	Symbol     string    `json:"symbol"`
	Name       string    `json:"name"`
	Amount     float64   `json:"amount"`
	ValueUSD   float64   `json:"valueUSD,omitempty"`
	PriceUSD   float64   `json:"priceUSD"`
	Allocation float64   `json:"allocation,omitempty"`
	APY        float64   `json:"apy,omitempty"`
	RiskScore  float64   `json:"riskScore,omitempty"`
	LastUpdate time.Time `json:"lastUpdate,omitempty"`
}

type Position struct {
	ID           string    `json:"id"`
	Asset        string    `json:"asset"`
	Type         string    `json:"type"`
	Size         float64   `json:"size"`
	EntryPrice   float64   `json:"entryPrice"`
	CurrentPrice float64   `json:"currentPrice,omitempty"`
	Pnl          float64   `json:"pnl,omitempty"`
	PnlPercent   float64   `json:"pnlPercent,omitempty"`
	Leverage     float64   `json:"leverage,omitempty"`
	Status       string    `json:"status,omitempty"`
	OpenedAt     time.Time `json:"openedAt,omitempty"`
	ClosedAt     time.Time `json:"closedAt,omitempty"`
}

// Placeholder handlers for unimplemented endpoints
func (s *Server) getPortfolioAssets(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, []Asset{})
}

func (s *Server) addAssetToPortfolio(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusCreated, Asset{})
}

func (s *Server) rebalancePortfolio(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"portfolioId": mux.Vars(r)["portfolioId"],
		"actions":     []interface{}{},
		"timestamp":   time.Now(),
	})
}

func (s *Server) getPortfolioRisk(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"portfolioId":     mux.Vars(r)["portfolioId"],
		"totalRiskScore":  0.3,
		"riskLevel":       "medium",
		"recommendations": []string{"Portfolio is well-diversified"},
		"assessmentTime":  time.Now(),
	})
}

func (s *Server) getMarketData(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, []interface{}{})
}

func (s *Server) listStrategies(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, []interface{}{})
}

func (s *Server) executeStrategy(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"strategyId": mux.Vars(r)["strategyId"],
		"executedAt": time.Now(),
		"success":    true,
	})
}

func (s *Server) listAgents(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, []interface{}{})
}

func (s *Server) executeAgent(w http.ResponseWriter, r *http.Request) {
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"agentId":      mux.Vars(r)["agentId"],
		"executedAt":   time.Now(),
		"success":      true,
		"actionsTaken": []string{"Task completed successfully"},
	})
}
