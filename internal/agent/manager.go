package agent

import (
	"fmt"

	"github.com/aegis-protocol/aegis-core/pkg/mcpclient"
)

type Manager struct {
	agents       []Agent
	hederaClient *mcpclient.HederaClient
}

type Agent struct {
	ID           string
	Name         string
	Description  string
	Status       string
	Capabilities []string
}

func NewManager() *Manager {
	hederaClient := mcpclient.NewHederaClient()

	// Discover Hedera agents
	hederaAgents, _ := hederaClient.DiscoverAgents()

	// Convert Hedera agents to our internal format
	agents := []Agent{
		{
			ID:           "arbitrage-1",
			Name:         "Cross-Chain Arbitrage",
			Description:  "Automated arbitrage detection across multiple chains",
			Status:       "Active",
			Capabilities: []string{"Price monitoring", "Gas optimization", "Cross-chain swaps"},
		},
		{
			ID:           "yield-1",
			Name:         "Yield Optimizer",
			Description:  "Automatically moves funds to highest yielding protocols",
			Status:       "Idle",
			Capabilities: []string{"APY comparison", "Risk assessment", "Auto-compounding"},
		},
	}

	// Add Hedera agents
	for _, hederaAgent := range hederaAgents {
		agents = append(agents, Agent{
			ID:           hederaAgent.ID,
			Name:         hederaAgent.Name,
			Description:  hederaAgent.Description,
			Status:       hederaAgent.Status,
			Capabilities: hederaAgent.Capabilities,
		})
	}

	return &Manager{
		agents:       agents,
		hederaClient: hederaClient,
	}
}

func (m *Manager) GetAgents() []Agent {
	return m.agents
}

func (m *Manager) StartAgent(id string) error {
	for i := range m.agents {
		if m.agents[i].ID == id {
			m.agents[i].Status = "Active"
			return nil
		}
	}
	return fmt.Errorf("agent not found: %s", id)
}

func (m *Manager) StopAgent(id string) error {
	for i := range m.agents {
		if m.agents[i].ID == id {
			m.agents[i].Status = "Idle"
			return nil
		}
	}
	return fmt.Errorf("agent not found: %s", id)
}

func (m *Manager) CoordinateHederaAgents(task string, requiredCapabilities []string) ([]Agent, error) {
	if m.hederaClient == nil {
		return nil, fmt.Errorf("hedera client not initialized")
	}

	hederaAgents, err := m.hederaClient.CoordinateAgents(task, requiredCapabilities)
	if err != nil {
		return nil, err
	}

	// Convert Hedera agents to our internal format
	var agents []Agent
	for _, hederaAgent := range hederaAgents {
		agents = append(agents, Agent{
			ID:           hederaAgent.ID,
			Name:         hederaAgent.Name,
			Description:  hederaAgent.Description,
			Status:       hederaAgent.Status,
			Capabilities: hederaAgent.Capabilities,
		})
	}

	return agents, nil
}

func (m *Manager) ExecuteHederaTask(agentID string, task string, parameters map[string]interface{}) (string, error) {
	if m.hederaClient == nil {
		return "", fmt.Errorf("hedera client not initialized")
	}

	return m.hederaClient.ExecuteAgentTask(agentID, task, parameters)
}

func (m *Manager) GetHederaMetrics() map[string]interface{} {
	if m.hederaClient == nil {
		return map[string]interface{}{"error": "hedera client not initialized"}
	}

	return m.hederaClient.GetAgentMetrics()
}
