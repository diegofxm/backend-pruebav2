package handler

import (
	"net/http"
	"secop-blockchain/internal/blockchain"
	"secop-blockchain/internal/service"

	"github.com/gin-gonic/gin"
)

// WorkflowHandler handles workflow-related HTTP requests
type WorkflowHandler struct {
	services *service.Services
}

// NewWorkflowHandler creates a new workflow handler
func NewWorkflowHandler(services *service.Services) *WorkflowHandler {
	return &WorkflowHandler{
		services: services,
	}
}

// GetSteps returns workflow steps
func (h *WorkflowHandler) GetSteps(c *gin.Context) {
	steps := h.services.Workflow.GetWorkflowSteps()
	c.JSON(http.StatusOK, gin.H{"steps": steps})
}

// GetContractStatus returns contract workflow status
func (h *WorkflowHandler) GetContractStatus(c *gin.Context) {
	contractID := c.Param("id")
	status, err := h.services.Workflow.GetWorkflowStatus(contractID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

// ValidateStep validates a workflow step
func (h *WorkflowHandler) ValidateStep(c *gin.Context) {
	contractID := c.Param("id")
	
	var req struct {
		StepNumber    int    `json:"step_number"`
		ValidatorID   string `json:"validator_id"`
		ValidatorName string `json:"validator_name"`
		Role          string `json:"role"`
		Approved      bool   `json:"approved"`
		Comments      string `json:"comments"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	role := blockchain.AdminRole(req.Role)
	err := h.services.Workflow.ValidateStep(contractID, req.StepNumber, req.ValidatorID, req.ValidatorName, role, req.Approved, req.Comments)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Paso validado exitosamente"})
}

// AddAudit adds audit observation
func (h *WorkflowHandler) AddAudit(c *gin.Context) {
	contractID := c.Param("id")
	
	var req struct {
		AuditorID   string `json:"auditor_id"`
		Role        string `json:"role"`
		Observation string `json:"observation"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	role := blockchain.AdminRole(req.Role)
	err := h.services.Workflow.AddAuditObservation(contractID, req.AuditorID, role, req.Observation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Observación de auditoría agregada"})
}