package blockchain

import (
	"errors"
	"fmt"
	"time"
	"secop-blockchain/internal/config"

	"github.com/google/uuid"
)

// WorkflowManager maneja el flujo de validación de contratos
type WorkflowManager struct {
	blockchain *Blockchain
}

// NewWorkflowManager crea un nuevo gestor de flujo de trabajo
func NewWorkflowManager(bc *Blockchain) *WorkflowManager {
	return &WorkflowManager{
		blockchain: bc,
	}
}

// GetWorkflowSteps define los pasos del flujo de trabajo SECOP
func (wm *WorkflowManager) GetWorkflowSteps() []WorkflowStep {
	return []WorkflowStep{
		{StepNumber: 1, Role: RoleProjectDeveloper, Name: "Creación del Proyecto", Required: true},
		{StepNumber: 2, Role: RoleTechnicalCommission, Name: "Revisión Técnica", Required: true},
		{StepNumber: 3, Role: RoleLegalCommission, Name: "Revisión Jurídica", Required: true},
		{StepNumber: 4, Role: RoleContractsChief, Name: "Aprobación Jefe de Contratos", Required: true},
		{StepNumber: 5, Role: RoleAdminChief, Name: "Aprobación Jefe Administrativo", Required: true},
		{StepNumber: 6, Role: RoleBudgetAuthority, Name: "Autorización Ordenador del Gasto", Required: true},
	}
}

// WorkflowStep representa un paso en el flujo de trabajo
type WorkflowStep struct {
	StepNumber int       `json:"step_number"`
	Role       AdminRole `json:"role"`
	Name       string    `json:"name"`
	Required   bool      `json:"required"`
}

// InitializeContractWorkflow inicializa el flujo de trabajo para un contrato
func (wm *WorkflowManager) InitializeContractWorkflow(contract *Contract) error {
	steps := wm.GetWorkflowSteps()
	contract.ValidationSteps = make([]ValidationStep, len(steps))
	
	for i, step := range steps {
		contract.ValidationSteps[i] = ValidationStep{
			StepNumber: step.StepNumber,
			Role:       step.Role,
			Status:     ValidationPending,
			Required:   step.Required,
			Timestamp:  time.Time{}, // Se establecerá cuando se valide
		}
	}
	
	contract.CurrentStep = 1
	contract.Status = StatusDraft
	contract.UpdatedAt = config.GetColombianTime()
	
	// Registrar en auditoría
	wm.addAuditEntry(contract, "WORKFLOW_INITIALIZED", contract.CreatedBy, RoleProjectDeveloper, "Flujo de trabajo inicializado")
	
	return nil
}

// ValidateStep valida un paso específico del flujo de trabajo
func (wm *WorkflowManager) ValidateStep(contractID string, stepNumber int, validatorID string, validatorName string, role AdminRole, approved bool, comments string) error {
	contract, exists := wm.blockchain.Contracts[contractID]
	if !exists {
		return errors.New("contrato no encontrado")
	}
	
	// Verificar que es el paso correcto
	if stepNumber != contract.CurrentStep {
		return fmt.Errorf("paso inválido. Paso actual: %d, paso solicitado: %d", contract.CurrentStep, stepNumber)
	}
	
	// Verificar que el paso existe
	if stepNumber > len(contract.ValidationSteps) {
		return errors.New("número de paso inválido")
	}
	
	// Obtener el paso actual
	step := &contract.ValidationSteps[stepNumber-1]
	
	// Actualizar el paso
	step.ValidatorID = validatorID
	step.ValidatorName = validatorName
	step.Timestamp = config.GetColombianTime()
	step.Comments = comments
	
	if approved {
		step.Status = ValidationApproved
		contract.CurrentStep++
		contract.Status = wm.getStatusForStep(contract.CurrentStep)
		wm.addAuditEntry(contract, "STEP_APPROVED", validatorID, role, fmt.Sprintf("Paso %d aprobado: %s", stepNumber, comments))
	} else {
		step.Status = ValidationRejected
		contract.Status = StatusRejected
		wm.addAuditEntry(contract, "STEP_REJECTED", validatorID, role, fmt.Sprintf("Paso %d rechazado: %s", stepNumber, comments))
	}
	
	contract.UpdatedAt = config.GetColombianTime()
	
	// Crear bloque para registrar la validación
	blockData := map[string]interface{}{
		"type":        "VALIDATION",
		"contract_id": contractID,
		"step":        stepNumber,
		"validator":   validatorID,
		"role":        string(role),
		"approved":    approved,
		"comments":    comments,
		"timestamp":   config.GetColombianTime(),
	}
	
	// Agregar bloque y obtener hash
	block, err := wm.blockchain.AddBlock(blockData)
	if err != nil {
		return err
	}

	// Actualizar audit trail con block hash
	if len(contract.AuditTrail) > 0 {
		contract.AuditTrail[len(contract.AuditTrail)-1].BlockHash = block.Hash
	}

	return nil
}

// getStatusForStep retorna el estado correspondiente al paso actual
func (wm *WorkflowManager) getStatusForStep(stepNumber int) ContractStatus {
	switch stepNumber {
	case 1:
		return StatusDraft
	case 2:
		return StatusTechnicalReview
	case 3:
		return StatusLegalReview
	case 4:
		return StatusContractsReview
	case 5:
		return StatusAdminReview
	case 6:
		return StatusBudgetReview
	default:
		return StatusAuthorizedForPublication
	}
}

// AddAuditObservation agrega una observación de auditoría (control externo)
func (wm *WorkflowManager) AddAuditObservation(contractID string, auditorID string, role AdminRole, observation string) error {
	contract, exists := wm.blockchain.Contracts[contractID]
	if !exists {
		return errors.New("contrato no encontrado")
	}
	
	// Verificar que es un rol de control externo
	if role != RoleComptroller && role != RoleProsecutor && role != RoleCitizen {
		return errors.New("rol no autorizado para auditoría")
	}
	
	// Agregar observación de auditoría
	auditEntry := AuditEntry{
		ID:          uuid.New().String(),
		Action:      "AUDIT_OBSERVATION",
		UserID:      auditorID,
		UserRole:    role,
		Timestamp:   config.GetColombianTime(),
		Description: observation,
		IPAddress:   "", // Se puede agregar desde el contexto HTTP
	}
	
	contract.AuditTrail = append(contract.AuditTrail, auditEntry)
	
	// Crear bloque para registrar la observación de auditoría
	blockData := map[string]interface{}{
		"type":        "AUDIT_OBSERVATION",
		"contract_id": contractID,
		"auditor":     auditorID,
		"role":        string(role),
		"observation": observation,
		"timestamp":   config.GetColombianTime(),
	}
	
	// Agregar bloque y obtener hash
	block, err := wm.blockchain.AddBlock(blockData)
	if err != nil {
		return err
	}
	
	// Actualizar audit trail con block hash
	contract.AuditTrail[len(contract.AuditTrail)-1].BlockHash = block.Hash
	
	return nil
}

// addAuditEntry agrega una entrada al registro de auditoría
func (wm *WorkflowManager) addAuditEntry(contract *Contract, action string, userID string, role AdminRole, description string) {
	entry := AuditEntry{
		ID:          uuid.New().String(),
		Action:      action,
		UserID:      userID,
		UserRole:    role,
		Timestamp:   config.GetColombianTime(),
		Description: description,
		IPAddress:   "", // Se puede agregar desde el contexto HTTP
	}
	
	contract.AuditTrail = append(contract.AuditTrail, entry)
}

// GetContractWorkflowStatus retorna el estado actual del flujo de trabajo
func (wm *WorkflowManager) GetContractWorkflowStatus(contractID string) (*WorkflowStatus, error) {
	contract, exists := wm.blockchain.Contracts[contractID]
	if !exists {
		return nil, errors.New("contrato no encontrado")
	}
	
	completedSteps := 0
	for _, step := range contract.ValidationSteps {
		if step.Status == ValidationApproved {
			completedSteps++
		}
	}
	
	return &WorkflowStatus{
		ContractID:     contractID,
		CurrentStep:    contract.CurrentStep,
		TotalSteps:     len(contract.ValidationSteps),
		CompletedSteps: completedSteps,
		Status:         contract.Status,
		CanAdvance:     contract.Status != StatusRejected && contract.Status != StatusCompleted,
		NextRole:       wm.getNextRole(contract),
	}, nil
}

// WorkflowStatus representa el estado del flujo de trabajo
type WorkflowStatus struct {
	ContractID     string         `json:"contract_id"`
	CurrentStep    int            `json:"current_step"`
	TotalSteps     int            `json:"total_steps"`
	CompletedSteps int            `json:"completed_steps"`
	Status         ContractStatus `json:"status"`
	CanAdvance     bool           `json:"can_advance"`
	NextRole       AdminRole      `json:"next_role"`
}

// getNextRole retorna el siguiente rol que debe validar
func (wm *WorkflowManager) getNextRole(contract *Contract) AdminRole {
	if contract.CurrentStep <= len(contract.ValidationSteps) {
		return contract.ValidationSteps[contract.CurrentStep-1].Role
	}
	return ""
}

// GetWorkflowStatus obtiene el estado actual del flujo de trabajo de un contrato
func (wm *WorkflowManager) GetWorkflowStatus(contractID string) (map[string]interface{}, error) {
	contract, exists := wm.blockchain.Contracts[contractID]
	if !exists {
		return nil, errors.New("contrato no encontrado")
	}

	// Calcular progreso
	completedSteps := 0
	totalSteps := len(contract.ValidationSteps)
	
	for _, step := range contract.ValidationSteps {
		if step.Status == ValidationApproved {
			completedSteps++
		}
	}

	progress := float64(completedSteps) / float64(totalSteps) * 100

	status := map[string]interface{}{
		"contract_id":      contractID,
		"current_step":     contract.CurrentStep,
		"total_steps":      totalSteps,
		"completed_steps":  completedSteps,
		"progress":         progress,
		"status":           string(contract.Status),
		"validation_steps": contract.ValidationSteps,
		"audit_trail":      contract.AuditTrail,
		"created_at":       contract.CreatedAt,
		"updated_at":       contract.UpdatedAt,
	}

	return status, nil
}