package handler

import (
	"net/http"
	"secop-blockchain/internal/service"
	"secop-blockchain/internal/config"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check and monitoring requests
type HealthHandler struct {
	services *service.Services
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(services *service.Services) *HealthHandler {
	return &HealthHandler{
		services: services,
	}
}

// GetHealth returns basic health status
func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": config.GetColombianTime(),
		"service":   "SECOP Blockchain API",
	})
}

// GetDetailedHealth returns detailed system health information
func (h *HealthHandler) GetDetailedHealth(c *gin.Context) {
	blockchainHeight := h.services.Blockchain.GetBlockchainHeight()
	peerCount := len(h.services.P2P.GetPeers())
	
	health := gin.H{
		"status":     "healthy",
		"timestamp":  config.GetColombianTime(),
		"service":    "SECOP Blockchain API",
		"blockchain": gin.H{
			"height":      blockchainHeight,
			"last_block":  h.services.Blockchain.GetLastBlockHash(),
			"is_synced":   h.services.P2P.IsSynced(),
		},
		"network": gin.H{
			"peer_count":     peerCount,
			"network_health": h.services.P2P.GetNetworkHealth(),
		},
		"system": gin.H{
			"uptime": time.Since(config.GetColombianTime()).String(), // This would be calculated from startup time
		},
	}
	
	c.JSON(http.StatusOK, health)
}

// GetReadiness returns readiness probe for Kubernetes/Docker
func (h *HealthHandler) GetReadiness(c *gin.Context) {
	// Check if blockchain is initialized and has at least genesis block
	if h.services.Blockchain.GetBlockchainHeight() < 1 {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"reason": "blockchain not initialized",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}

// GetLiveness returns liveness probe for Kubernetes/Docker
func (h *HealthHandler) GetLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Health returns comprehensive health information
func (h *HealthHandler) Health(c *gin.Context) {
	health := map[string]interface{}{
		"status":           "healthy",
		"timestamp":        config.GetColombianTime(),
		"blockchain":       h.services.Blockchain.GetNetworkHealth(),
		"p2p":             h.services.P2P.GetNetworkHealth(),
		"peers":           len(h.services.P2P.GetPeers()),
		"blockchain_height": h.services.Blockchain.GetBlockchainHeight(),
	}
	
	c.JSON(http.StatusOK, health)
}

// Stats returns system statistics
func (h *HealthHandler) Stats(c *gin.Context) {
	stats := map[string]interface{}{
		"blockchain_height":  h.services.Blockchain.GetBlockchainHeight(),
		"total_contracts":    len(h.services.Blockchain.GetAllContracts()),
		"total_peers":        len(h.services.P2P.GetPeers()),
		"chain_valid":        h.services.Blockchain.IsChainValid(),
		"is_synced":          h.services.P2P.IsSynced(),
		"uptime":            config.GetColombianTime(),
	}
	
	c.JSON(http.StatusOK, stats)
}

// GetBlocks returns blockchain blocks
func (h *HealthHandler) GetBlocks(c *gin.Context) {
	chain := h.services.Blockchain.GetChain()
	c.JSON(http.StatusOK, gin.H{
		"blocks": chain,
		"height": len(chain),
	})
}