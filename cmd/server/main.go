package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"secop-blockchain/internal/config"
	"secop-blockchain/internal/handler"
	"secop-blockchain/internal/service"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}
	
	// Load configuration
	cfg := config.Load()
	
	fmt.Printf("🚀 Iniciando SECOP Blockchain v2\n")
	fmt.Printf("📍 Nodo: %s\n", cfg.P2P.NodeID)
	fmt.Printf("🏛️ Entidad: %s\n", cfg.Entity.Type)
	fmt.Printf("🌐 Dirección: %s:%s\n", cfg.Server.Address, cfg.Server.Port)

	// Initialize services
	services := service.NewServices(cfg)
	
	// Setup bootstrap peers if configured
	setupBootstrapPeers(services, cfg)
	
	// System will start clean without example data
	if cfg.Entity.Type == "DNP" {
		createExampleContracts(services)
	}

	// Setup routes
	router := handler.SetupRoutes(cfg, services)
	
	// Start periodic tasks
	go startPeriodicTasks(services)

	fmt.Printf("✅ Servidor iniciado en puerto %s\n", cfg.Server.Port)
	fmt.Printf("🔗 API disponible en http://%s:%s/api/\n", cfg.Server.Address, cfg.Server.Port)
	
	// Start server
	if err := router.Run(":" + cfg.Server.Port); err != nil {
		log.Fatal("Error iniciando servidor:", err)
	}
}

func setupBootstrapPeers(services *service.Services, cfg *config.Config) {
	if len(cfg.P2P.BootstrapPeers) == 0 {
		fmt.Printf("🌐 Modo descubrimiento dinámico\n")
		return
	}
	
	fmt.Printf("🔗 Configurando %d peers bootstrap\n", len(cfg.P2P.BootstrapPeers))
	// TODO: Implement bootstrap peer setup logic
}

func createExampleContracts(services *service.Services) {
	// Function removed - system starts clean
	fmt.Printf("✅ Sistema iniciado sin datos de ejemplo\n")
}

func startPeriodicTasks(services *service.Services) {
	fmt.Printf("⏰ Iniciando tareas periódicas...\n")
	// TODO: Implement periodic sync and health checks
}