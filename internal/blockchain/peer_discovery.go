package blockchain

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	"secop-blockchain/internal/config"
)

// PeerDiscovery manages dynamic peer discovery for government entities
type PeerDiscovery struct {
	registryURL     string
	nodeID          string
	nodeAddress     string
	entityType      string
	knownPeers      map[string]*PeerInfo
	mutex           sync.RWMutex
	discoveryTicker *time.Ticker
}

// PeerInfo contains information about a discovered peer
type PeerInfo struct {
	ID          string    `json:"id"`
	Address     string    `json:"address"`
	Port        string    `json:"port"`
	EntityType  string    `json:"entity_type"`
	LastSeen    time.Time `json:"last_seen"`
	IsActive    bool      `json:"is_active"`
	PublicKey   string    `json:"public_key,omitempty"`
}

// EntityType defines the type of government entity
type EntityType string

const (
	EntityGovernment   EntityType = "GOVERNMENT"   // Entidades de Gobierno Nacional (DNP, etc.)
	EntityMunicipality EntityType = "MUNICIPALITY" // Municipio
	EntityDepartment   EntityType = "DEPARTMENT"   // Departamento
	EntityMinistry     EntityType = "MINISTRY"     // Ministerio
	EntityControl      EntityType = "CONTROL"      // Entidad de Control
	EntityDNP          EntityType = "DNP"          // Departamento Nacional de Planeación (legacy)
)

// NewPeerDiscovery creates a new peer discovery service
func NewPeerDiscovery(registryURL, nodeID, nodeAddress, entityType string) *PeerDiscovery {
	return &PeerDiscovery{
		registryURL: registryURL,
		nodeID:      nodeID,
		nodeAddress: nodeAddress,
		entityType:  entityType,
		knownPeers:  make(map[string]*PeerInfo),
	}
}

// Start begins the peer discovery process
func (pd *PeerDiscovery) Start() error {
	// Register this node with the discovery service
	if err := pd.registerNode(); err != nil {
		return fmt.Errorf("failed to register node: %v", err)
	}

	// Start periodic peer discovery
	pd.discoveryTicker = time.NewTicker(30 * time.Second)
	go pd.discoveryLoop()

	log.Printf("Peer discovery started for node %s (%s)", pd.nodeID, pd.entityType)
	return nil
}

// Stop stops the peer discovery process
func (pd *PeerDiscovery) Stop() {
	if pd.discoveryTicker != nil {
		pd.discoveryTicker.Stop()
	}
	
	// Unregister from discovery service
	pd.unregisterNode()
}

// registerNode registers this node with the central discovery service
func (pd *PeerDiscovery) registerNode() error {
	if pd.registryURL == "" {
		// If no registry URL, use bootstrap mode (for first nodes)
		log.Println("No registry URL configured, running in bootstrap mode")
		return nil
	}

	nodeInfo := PeerInfo{
		ID:         pd.nodeID,
		Address:    pd.nodeAddress,
		Port:       "", // Add empty port field
		EntityType: pd.entityType,
		LastSeen:   config.GetColombianTime(),
		IsActive:   true,
	}

	data, err := json.Marshal(nodeInfo)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/api/peers/register", pd.registryURL),
		"application/json",
		strings.NewReader(string(data)),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	return nil
}

// unregisterNode removes this node from the discovery service
func (pd *PeerDiscovery) unregisterNode() {
	if pd.registryURL == "" {
		return
	}

	url := fmt.Sprintf("%s/api/peers/unregister/%s", pd.registryURL, pd.nodeID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		log.Printf("Error creating unregister request: %v", err)
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error unregistering node: %v", err)
		return
	}
	defer resp.Body.Close()
}

// discoveryLoop periodically discovers new peers
func (pd *PeerDiscovery) discoveryLoop() {
	for range pd.discoveryTicker.C {
		if err := pd.discoverPeers(); err != nil {
			log.Printf("Peer discovery error: %v", err)
		}
	}
}

// discoverPeers fetches the list of active peers from the discovery service
func (pd *PeerDiscovery) discoverPeers() error {
	if pd.registryURL == "" {
		return nil // Bootstrap mode, no discovery needed
	}

	resp, err := http.Get(fmt.Sprintf("%s/api/peers", pd.registryURL))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var peers []PeerInfo
	if err := json.NewDecoder(resp.Body).Decode(&peers); err != nil {
		return err
	}

	pd.mutex.Lock()
	defer pd.mutex.Unlock()

	// Update known peers
	for _, peer := range peers {
		if peer.ID != pd.nodeID { // Don't add ourselves
			pd.knownPeers[peer.ID] = &peer
		}
	}

	// Remove inactive peers
	for id, peer := range pd.knownPeers {
		if config.GetColombianTime().Sub(peer.LastSeen) > 5*time.Minute {
			delete(pd.knownPeers, id)
		}
	}

	log.Printf("Discovered %d active peers", len(pd.knownPeers))
	return nil
}

// GetActivePeers returns a list of currently active peers
func (pd *PeerDiscovery) GetActivePeers() []*PeerInfo {
	pd.mutex.RLock()
	defer pd.mutex.RUnlock()

	peers := make([]*PeerInfo, 0, len(pd.knownPeers))
	for _, peer := range pd.knownPeers {
		if peer.IsActive {
			peers = append(peers, peer)
		}
	}

	return peers
}

// GetPeersByType returns peers of a specific entity type
func (pd *PeerDiscovery) GetPeersByType(entityType EntityType) []*PeerInfo {
	pd.mutex.RLock()
	defer pd.mutex.RUnlock()

	var peers []*PeerInfo
	for _, peer := range pd.knownPeers {
		if peer.EntityType == string(entityType) && peer.IsActive {
			peers = append(peers, peer)
		}
	}

	return peers
}

// AddBootstrapPeer manually adds a bootstrap peer (for initial network formation)
func (pd *PeerDiscovery) AddBootstrapPeer(id, address, entityType string) {
	pd.mutex.Lock()
	defer pd.mutex.Unlock()

	pd.knownPeers[id] = &PeerInfo{
		ID:         id,
		Address:    address,
		Port:       "", // Add empty port field
		EntityType: entityType,
		LastSeen:   config.GetColombianTime(),
		IsActive:   true,
	}

	log.Printf("Added bootstrap peer: %s (%s)", id, entityType)
}

// GetPeerCount returns the number of known active peers
func (pd *PeerDiscovery) GetPeerCount() int {
	pd.mutex.RLock()
	defer pd.mutex.RUnlock()

	count := 0
	for _, peer := range pd.knownPeers {
		if peer.IsActive {
			count++
		}
	}
	return count
}
