package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"secop-blockchain/internal/config"
)

// Peer representa un nodo peer en la red
type Peer struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Port     string `json:"port"`
	LastSeen time.Time `json:"last_seen"`
	Active   bool   `json:"active"`
}

// P2PNetwork maneja la comunicación entre nodos
type P2PNetwork struct {
	NodeID        string
	Address       string
	Port          string
	Peers         map[string]*Peer
	Blockchain    *Blockchain
	PeerDiscovery *PeerDiscovery
	mutex         sync.RWMutex
}

// NewP2PNetwork crea una nueva instancia de red P2P
func NewP2PNetwork(nodeID, address, port string, blockchain *Blockchain, discoveryRegistryURL, entityType string) *P2PNetwork {
	network := &P2PNetwork{
		NodeID:     nodeID,
		Address:    address,
		Port:       port,
		Peers:      make(map[string]*Peer),
		Blockchain: blockchain,
	}
	
	// Initialize peer discovery
	network.PeerDiscovery = NewPeerDiscovery(discoveryRegistryURL, nodeID, address, entityType)
	
	return network
}

// Start starts the P2P network
func (p2p *P2PNetwork) Start() error {
	// Start peer discovery
	if err := p2p.PeerDiscovery.Start(); err != nil {
		return fmt.Errorf("failed to start peer discovery: %v", err)
	}
	
	// Start periodic peer synchronization
	go p2p.syncPeersLoop()
	
	fmt.Printf("P2P network started for node %s", p2p.NodeID)
	return nil
}

// Stop stops the P2P network
func (p2p *P2PNetwork) Stop() {
	p2p.PeerDiscovery.Stop()
	fmt.Printf("P2P network stopped for node %s", p2p.NodeID)
}

// syncPeersLoop periodically synchronizes with discovered peers
func (p2p *P2PNetwork) syncPeersLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		p2p.syncWithDiscoveredPeers()
	}
}

// syncWithDiscoveredPeers synchronizes the peer list with discovered peers
func (p2p *P2PNetwork) syncWithDiscoveredPeers() {
	discoveredPeers := p2p.PeerDiscovery.GetActivePeers()
	
	p2p.mutex.Lock()
	defer p2p.mutex.Unlock()
	
	// Add new discovered peers
	for _, peerInfo := range discoveredPeers {
		if _, exists := p2p.Peers[peerInfo.ID]; !exists {
			peer := &Peer{
				ID:       peerInfo.ID,
				Address:  peerInfo.Address,
				Port:     peerInfo.Port,
				LastSeen: peerInfo.LastSeen,
				Active:   peerInfo.IsActive,
			}
			p2p.Peers[peerInfo.ID] = peer
			fmt.Printf("Added discovered peer: %s (%s:%s)", peerInfo.ID, peerInfo.Address, peerInfo.Port)
		}
	}
	
	// Remove peers that are no longer discovered
	discoveredIDs := make(map[string]bool)
	for _, peerInfo := range discoveredPeers {
		discoveredIDs[peerInfo.ID] = true
	}
	
	for id := range p2p.Peers {
		if !discoveredIDs[id] {
			delete(p2p.Peers, id)
			fmt.Printf("Removed inactive peer: %s", id)
		}
	}
}

// AddBootstrapPeer adds a bootstrap peer for initial network formation
func (p2p *P2PNetwork) AddBootstrapPeer(id, address, entityType string) {
	p2p.PeerDiscovery.AddBootstrapPeer(id, address, entityType)
	
	p2p.mutex.Lock()
	defer p2p.mutex.Unlock()
	
	p2p.Peers[id] = &Peer{
		ID:       id,
		Address:  address,
		Port:     "", // Will be extracted from address if needed
		LastSeen: config.GetColombianTime(),
		Active:   true,
	}
}

// AddPeer agrega un nuevo peer a la red
func (p2p *P2PNetwork) AddPeer(peerID, address, port string) error {
	p2p.mutex.Lock()
	defer p2p.mutex.Unlock()
	
	if _, exists := p2p.Peers[peerID]; exists {
		return fmt.Errorf("peer %s already exists", peerID)
	}
	
	p2p.Peers[peerID] = &Peer{
		ID:       peerID,
		Address:  address,
		Port:     port,
		LastSeen: config.GetColombianTime(),
		Active:   true,
	}
	
	fmt.Printf("🔗 Peer agregado: %s (%s:%s)\n", peerID, address, port)
	return nil
}

// BroadcastBlock envía un nuevo bloque a todos los peers
func (p2p *P2PNetwork) BroadcastBlock(block Block) {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	fmt.Printf("📡 Broadcasting bloque %s a %d peers\n", block.Hash, len(p2p.Peers))
	
	for peerID, peer := range p2p.Peers {
		if !peer.Active {
			continue
		}
		
		go func(peerID string, peer *Peer) {
			err := p2p.sendBlockToPeer(peer, block)
			if err != nil {
				fmt.Printf("❌ Error enviando bloque a %s: %v\n", peerID, err)
				p2p.markPeerInactive(peerID)
			} else {
				fmt.Printf("✅ Bloque enviado a %s\n", peerID)
			}
		}(peerID, peer)
	}
}

// sendBlockToPeer envía un bloque a un peer específico
func (p2p *P2PNetwork) sendBlockToPeer(peer *Peer, block Block) error {
	url := fmt.Sprintf("http://%s:%s/api/p2p/receive-block", peer.Address, peer.Port)
	
	blockData, err := json.Marshal(block)
	if err != nil {
		return err
	}
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(blockData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("peer respondió con status %d", resp.StatusCode)
	}
	
	return nil
}

// ReceiveBlock procesa un bloque recibido de otro peer
func (p2p *P2PNetwork) ReceiveBlock(block Block) error {
	fmt.Printf("📥 Bloque recibido de peer: %s\n", block.Hash)
	
	// Validar el bloque
	if !p2p.Blockchain.IsValidBlock(block) {
		return fmt.Errorf("bloque inválido recibido")
	}
	
	// Verificar si ya tenemos este bloque
	if p2p.Blockchain.HasBlock(block.Hash) {
		fmt.Printf("⚠️ Bloque %s ya existe, ignorando\n", block.Hash)
		return nil
	}
	
	// Agregar el bloque a nuestra cadena
	blockData := map[string]interface{}{
		"type":          block.Type,
		"data":          block.Data,
		"timestamp":     block.Timestamp,
		"previous_hash": block.PreviousHash,
		"nonce":         block.Nonce,
	}
	
	_, err := p2p.Blockchain.AddBlock(blockData)
	if err != nil {
		return fmt.Errorf("error agregando bloque: %v", err)
	}
	
	fmt.Printf("✅ Bloque %s agregado exitosamente\n", block.Hash)
	return nil
}

// SyncWithPeers sincroniza la blockchain con todos los peers
func (p2p *P2PNetwork) SyncWithPeers() error {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	fmt.Printf("🔄 Iniciando sincronización con %d peers\n", len(p2p.Peers))
	
	for peerID, peer := range p2p.Peers {
		if !peer.Active {
			continue
		}
		
		chain, err := p2p.requestChainFromPeer(peer)
		if err != nil {
			fmt.Printf("❌ Error obteniendo cadena de %s: %v\n", peerID, err)
			continue
		}
		
		// Si el peer tiene una cadena más larga y válida, la adoptamos
		if len(chain) > len(p2p.Blockchain.Chain) && p2p.Blockchain.IsValidChain(chain) {
			fmt.Printf("🔄 Adoptando cadena más larga de %s (%d bloques)\n", peerID, len(chain))
			// Convertir []Block a []*Block
			p2p.Blockchain.Chain = make([]*Block, len(chain))
			for i, block := range chain {
				blockCopy := block
				p2p.Blockchain.Chain[i] = &blockCopy
			}
			p2p.rebuildContractsFromChain()
		}
	}
	
	return nil
}

// requestChainFromPeer solicita la blockchain completa de un peer
func (p2p *P2PNetwork) requestChainFromPeer(peer *Peer) ([]Block, error) {
	url := fmt.Sprintf("http://%s:%s/api/p2p/get-chain", peer.Address, peer.Port)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer respondió con status %d", resp.StatusCode)
	}
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	var response struct {
		Chain []Block `json:"chain"`
	}
	
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	
	return response.Chain, nil
}

// rebuildContractsFromChain reconstruye el mapa de contratos desde la cadena
func (p2p *P2PNetwork) rebuildContractsFromChain() {
	p2p.Blockchain.Contracts = make(map[string]*Contract)
	
	for _, block := range p2p.Blockchain.Chain {
		if block.Type == "CONTRACT_CREATION" {
			var contract Contract
			err := json.Unmarshal([]byte(fmt.Sprintf("%v", block.Data)), &contract)
			if err == nil {
				p2p.Blockchain.Contracts[contract.ID] = &contract
			}
		}
	}
	
	fmt.Printf("🔄 Contratos reconstruidos: %d\n", len(p2p.Blockchain.Contracts))
}

// markPeerInactive marca un peer como inactivo
func (p2p *P2PNetwork) markPeerInactive(peerID string) {
	p2p.mutex.Lock()
	defer p2p.mutex.Unlock()
	
	if peer, exists := p2p.Peers[peerID]; exists {
		peer.Active = false
		fmt.Printf("⚠️ Peer %s marcado como inactivo\n", peerID)
	}
}

// GetActivePeers retorna la lista de peers activos
func (p2p *P2PNetwork) GetActivePeers() []*Peer {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	var activePeers []*Peer
	for _, peer := range p2p.Peers {
		if peer.Active {
			activePeers = append(activePeers, peer)
		}
	}
	
	return activePeers
}

// HealthCheck verifica el estado de todos los peers
func (p2p *P2PNetwork) HealthCheck() {
	p2p.mutex.Lock()
	defer p2p.mutex.Unlock()
	
	for peerID, peer := range p2p.Peers {
		url := fmt.Sprintf("http://%s:%s/api/health", peer.Address, peer.Port)
		
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url)
		
		if err != nil || resp.StatusCode != http.StatusOK {
			peer.Active = false
			fmt.Printf("💔 Peer %s no responde\n", peerID)
		} else {
			peer.Active = true
			peer.LastSeen = config.GetColombianTime()
			fmt.Printf("💚 Peer %s activo\n", peerID)
		}
		
		if resp != nil {
			resp.Body.Close()
		}
	}
}

// GetNetworkHealth returns the health status of the P2P network
func (p2p *P2PNetwork) GetNetworkHealth() map[string]interface{} {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	activePeers := 0
	totalPeers := len(p2p.Peers)
	
	// Count active peers (seen in last 5 minutes)
	fiveMinutesAgo := config.GetColombianTime().Add(-5 * time.Minute)
	for _, peer := range p2p.Peers {
		if peer.LastSeen.After(fiveMinutesAgo) {
			activePeers++
		}
	}
	
	// Get blockchain health
	blockchainHealth := p2p.Blockchain.GetNetworkHealth()
	
	health := map[string]interface{}{
		"node_id":           p2p.NodeID,
		"address":           fmt.Sprintf("%s:%d", p2p.Address, p2p.Port),
		"total_peers":       totalPeers,
		"active_peers":      activePeers,
		"peer_discovery":    p2p.PeerDiscovery != nil,
		"blockchain_health": blockchainHealth,
		"timestamp":         config.GetColombianTime(),
	}
	
	// Add peer details
	peerDetails := make(map[string]interface{})
	for id, peer := range p2p.Peers {
		peerDetails[id] = map[string]interface{}{
			"address":   peer.Address,
			"port":      peer.Port,
			"last_seen": peer.LastSeen,
			"active":    peer.LastSeen.After(fiveMinutesAgo),
		}
	}
	health["peers"] = peerDetails
	
	return health
}

// GetPeers returns all peers in the network
func (p2p *P2PNetwork) GetPeers() map[string]*Peer {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	peers := make(map[string]*Peer)
	for id, peer := range p2p.Peers {
		peers[id] = peer
	}
	return peers
}

// IsSynced checks if the P2P network is synchronized
func (p2p *P2PNetwork) IsSynced() bool {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	// Consider synced if we have active peers and blockchain is synced
	activePeers := 0
	fiveMinutesAgo := config.GetColombianTime().Add(-5 * time.Minute)
	for _, peer := range p2p.Peers {
		if peer.LastSeen.After(fiveMinutesAgo) {
			activePeers++
		}
	}
	
	return activePeers > 0 && p2p.Blockchain.IsSynced()
}

// RemovePeer removes a peer from the network
func (p2p *P2PNetwork) RemovePeer(id string) error {
	p2p.mutex.Lock()
	defer p2p.mutex.Unlock()
	
	if _, exists := p2p.Peers[id]; !exists {
		return fmt.Errorf("peer %s not found", id)
	}
	
	delete(p2p.Peers, id)
	fmt.Printf("❌ Peer %s eliminado\n", id)
	return nil
}

// SyncBlockchain synchronizes the blockchain with peers
func (p2p *P2PNetwork) SyncBlockchain() error {
	p2p.mutex.RLock()
	defer p2p.mutex.RUnlock()
	
	if len(p2p.Peers) == 0 {
		return fmt.Errorf("no peers available for synchronization")
	}
	
	// Simple sync implementation - in production this would be more sophisticated
	fmt.Printf("🔄 Sincronizando blockchain con %d peers\n", len(p2p.Peers))
	
	for peerID, peer := range p2p.Peers {
		// Check if peer is active
		fiveMinutesAgo := config.GetColombianTime().Add(-5 * time.Minute)
		if !peer.LastSeen.After(fiveMinutesAgo) {
			continue
		}
		
		fmt.Printf("📡 Sincronizando con peer %s (%s:%s)\n", peerID, peer.Address, peer.Port)
		// In a real implementation, this would fetch and compare blockchain data
	}
	
	return nil
}