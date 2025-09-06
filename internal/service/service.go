package service

import (
	"secop-blockchain/internal/blockchain"
	"secop-blockchain/internal/config"
)

// Services holds all business logic services
type Services struct {
	Blockchain *blockchain.Blockchain
	P2P        *blockchain.P2PNetwork
	Workflow   *blockchain.WorkflowManager
	Config     *config.Config
}

// NewServices creates and initializes all services
func NewServices(cfg *config.Config) *Services {
	// Initialize blockchain
	bc := blockchain.NewBlockchain()
	
	// Initialize P2P network
	p2pNetwork := blockchain.NewP2PNetwork(
		cfg.P2P.NodeID,
		cfg.Server.Address,
		cfg.Server.Port,
		bc,
		cfg.P2P.DiscoveryRegistryURL,
		cfg.Entity.Type,
	)
	
	// Initialize workflow manager
	workflowManager := blockchain.NewWorkflowManager(bc)
	
	return &Services{
		Blockchain: bc,
		P2P:        p2pNetwork,
		Workflow:   workflowManager,
		Config:     cfg,
	}
}