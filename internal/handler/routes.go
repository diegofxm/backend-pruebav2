package handler

import (
	"secop-blockchain/internal/config"
	"secop-blockchain/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all application routes
func SetupRoutes(cfg *config.Config, services *service.Services) *gin.Engine {
	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)
	
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
	}))

	// Initialize handlers
	contractHandler := NewContractHandler(services)
	workflowHandler := NewWorkflowHandler(services)
	p2pHandler := NewP2PHandler(services)
	healthHandler := NewHealthHandler(services)

	// API Routes
	api := r.Group("/api")
	{
		// Contract routes
		contracts := api.Group("/contracts")
		{
			contracts.GET("", contractHandler.GetAll)
			contracts.POST("", contractHandler.Create)
			contracts.POST("/validate", contractHandler.Validate)
			contracts.GET("/by-status/:status", contractHandler.GetByStatus)
			contracts.GET("/by-role/:role", contractHandler.GetByRole)
		}

		// Workflow routes
		workflow := api.Group("/workflow")
		{
			workflow.GET("/steps", workflowHandler.GetSteps)
		}

		// Contract workflow routes
		api.GET("/contracts/:id/workflow", workflowHandler.GetContractStatus)
		api.POST("/contracts/:id/validate-step", workflowHandler.ValidateStep)
		api.POST("/contracts/:id/audit", workflowHandler.AddAudit)

		// P2P routes
		p2p := api.Group("/p2p")
		{
			p2p.GET("/peers", p2pHandler.GetPeers)
			p2p.POST("/add-peer", p2pHandler.AddPeer)
			p2p.GET("/get-chain", p2pHandler.GetChain)
			p2p.POST("/receive-block", p2pHandler.ReceiveBlock)
			p2p.POST("/sync", p2pHandler.Sync)
		}

		// Health and stats routes
		api.GET("/health", healthHandler.Health)
		api.GET("/stats", healthHandler.Stats)
		api.GET("/blocks", healthHandler.GetBlocks)
	}

	return r
}