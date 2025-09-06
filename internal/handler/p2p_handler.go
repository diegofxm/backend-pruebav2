package handler

import (
	"net/http"
	"secop-blockchain/internal/blockchain"
	"secop-blockchain/internal/service"

	"github.com/gin-gonic/gin"
)

// P2PHandler handles P2P network-related HTTP requests
type P2PHandler struct {
	services *service.Services
}

// NewP2PHandler creates a new P2P handler
func NewP2PHandler(services *service.Services) *P2PHandler {
	return &P2PHandler{
		services: services,
	}
}

// GetPeers returns all connected peers
func (h *P2PHandler) GetPeers(c *gin.Context) {
	peers := h.services.P2P.GetPeers()
	c.JSON(http.StatusOK, gin.H{"peers": peers})
}

// AddPeer adds a new peer to the network
func (h *P2PHandler) AddPeer(c *gin.Context) {
	var req struct {
		ID      string `json:"id"`
		Address string `json:"address"`
		Port    string `json:"port"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.P2P.AddPeer(req.ID, req.Address, req.Port)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Peer agregado exitosamente"})
}

// RemovePeer removes a peer from the network
func (h *P2PHandler) RemovePeer(c *gin.Context) {
	peerID := c.Param("id")
	
	err := h.services.P2P.RemovePeer(peerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Peer eliminado exitosamente"})
}

// SyncBlockchain synchronizes blockchain with peers
func (h *P2PHandler) SyncBlockchain(c *gin.Context) {
	err := h.services.P2P.SyncBlockchain()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Sincronización completada"})
}

// BroadcastBlock broadcasts a block to all peers
func (h *P2PHandler) BroadcastBlock(c *gin.Context) {
	blockHash := c.Param("hash")
	
	// Find the block by hash in the blockchain
	var targetBlock *blockchain.Block
	for _, block := range h.services.Blockchain.Chain {
		if block.Hash == blockHash {
			targetBlock = block
			break
		}
	}
	
	if targetBlock == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bloque no encontrado"})
		return
	}
	
	// Broadcast the actual block object
	h.services.P2P.BroadcastBlock(*targetBlock)
	
	c.JSON(http.StatusOK, gin.H{"message": "Bloque transmitido a todos los peers"})
}

// GetNetworkHealth returns network health status
func (h *P2PHandler) GetNetworkHealth(c *gin.Context) {
	health := h.services.P2P.GetNetworkHealth()
	c.JSON(http.StatusOK, health)
}

// GetChain returns the blockchain
func (h *P2PHandler) GetChain(c *gin.Context) {
	chain := h.services.Blockchain.GetChain()
	c.JSON(http.StatusOK, gin.H{
		"chain":  chain,
		"height": len(chain),
	})
}

// ReceiveBlock receives a block from another peer
func (h *P2PHandler) ReceiveBlock(c *gin.Context) {
	var block blockchain.Block
	if err := c.ShouldBindJSON(&block); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate and add the block
	if !h.services.Blockchain.IsValidBlock(block) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bloque inválido"})
		return
	}

	// Check if we already have this block
	if h.services.Blockchain.HasBlock(block.Hash) {
		c.JSON(http.StatusOK, gin.H{"message": "Bloque ya existe"})
		return
	}

	// Add block to blockchain
	blockData := block.Data
	if blockData == nil {
		blockData = make(map[string]interface{})
	}
	
	_, err := h.services.Blockchain.AddBlock(blockData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bloque recibido y agregado"})
}

// Sync synchronizes the blockchain with peers
func (h *P2PHandler) Sync(c *gin.Context) {
	err := h.services.P2P.SyncBlockchain()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Sincronización completada"})
}