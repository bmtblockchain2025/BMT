package network

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"
)

// Peer represents a peer in the P2P network.
type Peer struct {
	ID         string
	Address    string
	Conn       net.Conn
	Latency    time.Duration
	Trusted    bool
	LastActive time.Time // Last time the peer was active
	Violation  int        // Count of protocol violations
}

// P2PNetwork represents the P2P network for the blockchain.
type P2PNetwork struct {
	NodeID        string
	Address       string
	Peers         map[string]*Peer
	mutex         sync.Mutex
	TLSConfig     *tls.Config
	listener      net.Listener
	discoveryChan chan *Peer // Channel for peer discovery
	workerPool    chan struct{}
	maxPeers      int
	blacklist     map[string]struct{} // Blacklisted peers
}

// NewP2PNetwork initializes a new P2P network.
func NewP2PNetwork(nodeID, address string, tlsConfig *tls.Config) *P2PNetwork {
	return &P2PNetwork{
		NodeID:        nodeID,
		Address:       address,
		Peers:         make(map[string]*Peer),
		TLSConfig:     tlsConfig,
		discoveryChan: make(chan *Peer, 10),
		workerPool:    make(chan struct{}, 10), // Limit concurrent workers to 10
		maxPeers:      50,                      // Limit maximum peers to 50
		blacklist:     make(map[string]struct{}),
	}
}

// Start starts the P2P network and listens for incoming connections.
func (p *P2PNetwork) Start() error {
	listener, err := tls.Listen("tcp", p.Address, p.TLSConfig)
	if err != nil {
		return err
	}
	p.listener = listener
	log.Printf("Node %s started P2P network at %s", p.NodeID, p.Address)

	go p.acceptConnections()
	return nil
}

// acceptConnections handles incoming connections from peers.
func (p *P2PNetwork) acceptConnections() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go p.handleNewPeer(conn)
	}
}

// handleNewPeer handles a new incoming peer connection.
func (p *P2PNetwork) handleNewPeer(conn net.Conn) {
	peerID := conn.RemoteAddr().String()
	if _, blacklisted := p.blacklist[peerID]; blacklisted {
		log.Printf("Rejected connection from blacklisted peer %s", peerID)
		conn.Close()
		return
	}
	if len(p.Peers) >= p.maxPeers {
		log.Printf("Max peers reached, rejecting connection from %s", peerID)
		conn.Close()
		return
	}

	peer := &Peer{
		ID:         peerID,
		Address:    conn.RemoteAddr().String(),
		Conn:       conn,
		Trusted:    false, // Default to untrusted
		LastActive: time.Now(),
	}

	p.mutex.Lock()
	p.Peers[peerID] = peer
	p.mutex.Unlock()

	log.Printf("Node %s connected to new peer %s", p.NodeID, peerID)
	go p.handlePeerCommunication(peer)
}

// handlePeerCommunication handles communication with a connected peer.
func (p *P2PNetwork) handlePeerCommunication(peer *Peer) {
	buffer := make([]byte, 1024)
	for {
		n, err := peer.Conn.Read(buffer)
		if err != nil {
			log.Printf("Error reading from peer %s: %v", peer.ID, err)
			p.DisconnectPeer(peer.ID)
			return
		}
		message := string(buffer[:n])
		log.Printf("Received message from peer %s: %s", peer.ID, message)
		peer.LastActive = time.Now()
		// Here you can handle different types of messages (e.g., transactions, blocks)
	}
}

// ConnectPeer connects to a new peer.
func (p *P2PNetwork) ConnectPeer(address string) error {
	select {
	case p.workerPool <- struct{}{}:
		go func() {
			defer func() { <-p.workerPool }()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			conn, err := tls.Dial("tcp", address, p.TLSConfig)
			if err != nil {
				log.Printf("Failed to connect to peer %s: %v", address, err)
				return
			}

			peer := &Peer{
				ID:         address,
				Address:    address,
				Conn:       conn,
				Trusted:    false,
				LastActive: time.Now(),
			}

			p.mutex.Lock()
			p.Peers[address] = peer
			p.mutex.Unlock()

			log.Printf("Node %s connected to peer %s", p.NodeID, address)
			p.measureLatency(peer)
			go p.handlePeerCommunication(peer)
		}()
	default:
		log.Printf("Worker pool is full, dropping connection request to %s", address)
	}
	return nil
}

// DisconnectPeer disconnects from a peer.
func (p *P2PNetwork) DisconnectPeer(peerID string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	peer, exists := p.Peers[peerID]
	if !exists {
		log.Printf("Peer %s not found", peerID)
		return
	}

	peer.Conn.Close()
	delete(p.Peers, peerID)
	log.Printf("Node %s disconnected from peer %s", p.NodeID, peerID)
}

// BroadcastMessage broadcasts a compressed message to all connected peers.
func (p *P2PNetwork) BroadcastMessage(message string) {
	compressedMessage, err := compressData([]byte(message))
	if err != nil {
		log.Printf("Failed to compress message: %v", err)
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, peer := range p.Peers {
		go func(peer *Peer) {
			_, err := peer.Conn.Write(compressedMessage)
			if err != nil {
				log.Printf("Failed to send message to peer %s: %v", peer.ID, err)
				peer.Violation++
				if peer.Violation > 3 {
					log.Printf("Peer %s added to blacklist due to violations", peer.ID)
					p.blacklistPeer(peer.ID)
				}
			}
		}(peer)
	}
}

// blacklistPeer adds a peer to the blacklist.
func (p *P2PNetwork) blacklistPeer(peerID string) {
	p.blacklist[peerID] = struct{}{}
	p.DisconnectPeer(peerID)
}

// compressData compresses data using gzip.
func compressData(data []byte) ([]byte, error) {
	var buf []byte
	writer := gzip.NewWriter(&buf)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return buf, nil
}

// DiscoverPeers listens for discovered peers and attempts to connect to them.
func (p *P2PNetwork) DiscoverPeers() {
	go func() {
		for peer := range p.discoveryChan {
			if err := p.ConnectPeer(peer.Address); err != nil {
				log.Printf("Failed to connect to discovered peer %s: %v", peer.Address, err)
			}
		}
	}()
}

// AddDiscoveredPeer adds a discovered peer to the discovery channel.
func (p *P2PNetwork) AddDiscoveredPeer(peer *Peer) {
	select {
	case p.discoveryChan <- peer:
		log.Printf("Discovered new peer: %s", peer.Address)
	default:
		log.Printf("Discovery channel full, dropping peer: %s", peer.Address)
	}
}

// measureLatency measures the latency to a peer.
func (p *P2PNetwork) measureLatency(peer *Peer) {
	start := time.Now()
	_, err := peer.Conn.Write([]byte("ping"))
	if err != nil {
		log.Printf("Failed to measure latency to peer %s: %v", peer.ID, err)
		return
	}
	peer.Latency = time.Since(start)
	log.Printf("Latency to peer %s: %v", peer.ID, peer.Latency)
}

// Stop stops the P2P network.
func (p *P2PNetwork) Stop() {
	p.listener.Close()
	log.Printf("Node %s stopped P2P network", p.NodeID)
}
