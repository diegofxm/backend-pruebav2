package handler

import (
	"net/http"
	"secop-blockchain/internal/blockchain"
	"secop-blockchain/internal/service"

	"github.com/gin-gonic/gin"
)

// ContractHandler handles contract-related HTTP requests
type ContractHandler struct {
	services *service.Services
}

// NewContractHandler creates a new contract handler
func NewContractHandler(services *service.Services) *ContractHandler {
	return &ContractHandler{
		services: services,
	}
}

// GetAll returns all contracts
func (h *ContractHandler) GetAll(c *gin.Context) {
	contracts := h.services.Blockchain.GetAllContracts()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(contracts),
		"data":    contracts,
	})
}

// Create creates a new contract
func (h *ContractHandler) Create(c *gin.Context) {
	var contract blockchain.Contract
	if err := c.ShouldBindJSON(&contract); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.Blockchain.AddContract(&contract)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast new block to peers
	if len(h.services.Blockchain.Chain) > 0 {
		lastBlock := *h.services.Blockchain.Chain[len(h.services.Blockchain.Chain)-1]
		go h.services.P2P.BroadcastBlock(lastBlock)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":     true,
		"message":     "Contrato creado exitosamente",
		"contract_id": contract.ID,
	})
}

// Validate validates a contract
func (h *ContractHandler) Validate(c *gin.Context) {
	var req struct {
		ContractID string `json:"contractId"`
		NodeID     string `json:"nodeId"`
		Approved   bool   `json:"approved"`
		Reason     string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.services.Blockchain.ValidateContract(req.ContractID, req.NodeID, req.Approved, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast validation block to peers
	if len(h.services.Blockchain.Chain) > 0 {
		lastBlock := *h.services.Blockchain.Chain[len(h.services.Blockchain.Chain)-1]
		go h.services.P2P.BroadcastBlock(lastBlock)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Validación registrada exitosamente",
	})
}

// GetByStatus returns contracts by status
func (h *ContractHandler) GetByStatus(c *gin.Context) {
	status := c.Param("status")
	contracts := h.services.Blockchain.GetContractsByStatus(blockchain.ContractStatus(status))
	c.JSON(http.StatusOK, gin.H{"contracts": contracts})
}

// GetByRole returns contracts by role
func (h *ContractHandler) GetByRole(c *gin.Context) {
	role := c.Param("role")
	contracts := h.services.Blockchain.GetContractsByRole(blockchain.AdminRole(role))
	c.JSON(http.StatusOK, gin.H{"contracts": contracts})
}