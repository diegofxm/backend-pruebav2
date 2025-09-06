package config

import (
	"os"
	"strings"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Server     ServerConfig
	Blockchain BlockchainConfig
	P2P        P2PConfig
	Entity     EntityConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port    string
	Address string
	Mode    string // gin mode: debug, release, test
}

// BlockchainConfig holds blockchain configuration
type BlockchainConfig struct {
	GenesisBlock bool
	Difficulty   int
}

// P2PConfig holds P2P network configuration
type P2PConfig struct {
	NodeID                string
	DiscoveryRegistryURL  string
	BootstrapPeers        []string
}

// EntityConfig holds entity-specific configuration
type EntityConfig struct {
	Type string // DNP, MUNICIPALITY, DEPARTMENT, MINISTRY, CONTROL
}

// ColombianTimezone represents Colombia's timezone (UTC-5)
var ColombianTimezone *time.Location

func init() {
	// Initialize Colombian timezone (UTC-5)
	var err error
	ColombianTimezone, err = time.LoadLocation("America/Bogota")
	if err != nil {
		// Fallback to fixed offset if timezone data is not available
		ColombianTimezone = time.FixedZone("COT", -5*60*60) // UTC-5
	}
}

// GetColombianTime returns current time in Colombian timezone
func GetColombianTime() time.Time {
	return time.Now().In(ColombianTimezone)
}

// ToColombianTime converts any time to Colombian timezone
func ToColombianTime(t time.Time) time.Time {
	return t.In(ColombianTimezone)
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnv("NODE_PORT", "8080"),
			Address: getEnv("NODE_ADDRESS", "localhost"),
			Mode:    getEnv("GIN_MODE", "debug"),
		},
		Blockchain: BlockchainConfig{
			GenesisBlock: getEnv("GENESIS_BLOCK", "false") == "true",
			Difficulty:   1,
		},
		P2P: P2PConfig{
			NodeID:               getEnv("NODE_ID", "secop-government-central-bogota"),
			DiscoveryRegistryURL: getEnv("PEER_DISCOVERY_REGISTRY_URL", ""),
			BootstrapPeers:       parseBootstrapPeers(getEnv("BOOTSTRAP_PEERS", "")),
		},
		Entity: EntityConfig{
			Type: getEnv("ENTITY_TYPE", "GOVERNMENT"),
		},
	}
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseBootstrapPeers parses bootstrap peers from environment variable
// Format: nodeId1:address1,nodeId2:address2
func parseBootstrapPeers(peersStr string) []string {
	if peersStr == "" {
		return []string{}
	}
	
	peers := strings.Split(peersStr, ",")
	var result []string
	
	for _, peer := range peers {
		peer = strings.TrimSpace(peer)
		if peer != "" {
			result = append(result, peer)
		}
	}
	
	return result
}