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

	"blockchain"
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
	Blockchain    *blockchain.Blockchain
}

// NewP2PNetwork initializes a new P2P network.
func NewP2PNetwork(nodeID, address string, tlsConfig *tls.Config, blockchain *blockchain.Blockchain) *P2PNetwork {
	return &P2PNetwork{
		NodeID:        nodeID,
		Address:       address,
		Peers:         make(map[string]*Peer),
		TLSConfig:     tlsConfig,
		discoveryChan: make(chan *Peer, 10),
		workerPool:    make(chan struct{}, 10), // Limit concurrent workers to 10
		maxPeers:      50,                      // Limit maximum peers to 50
		blacklist:     make(map[string]struct{}),
		Blockchain:    blockchain,
	}
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
		message := buffer[:n]
		p.handleMessage(message)
		peer.LastActive = time.Now()
	}
}

// handleMessage handles incoming messages from peers.
func (p *P2PNetwork) handleMessage(message []byte) {
	var msg map[string]interface{}
	err := json.Unmarshal(message, &msg)
	if err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	msgType, ok := msg["type"].(string)
	if !ok {
		log.Printf("Invalid message format")
		return
	}

	switch msgType {
	case "transaction":
		p.handleTransactionMessage(message)
	case "block":
		p.handleBlockMessage(message)
	default:
		log.Printf("Unknown message type: %s", msgType)
	}
}

// handleTransactionMessage handles an incoming transaction message.
func (p *P2PNetwork) handleTransactionMessage(txData []byte) {
	var tx blockchain.Transaction
	err := json.Unmarshal(txData, &tx)
	if err != nil {
		log.Printf("Failed to unmarshal transaction: %v", err)
		return
	}

	err = p.Blockchain.AddTransaction(&tx, p.NodeID)
	if err != nil {
		log.Printf("Failed to add transaction from peer: %v", err)
		return
	}

	log.Printf("Successfully added transaction from peer")
	p.BroadcastMessage(string(txData))
}

// handleBlockMessage handles an incoming block message.
func (p *P2PNetwork) handleBlockMessage(blockData []byte) {
	var block blockchain.MainBlock
	err := json.Unmarshal(blockData, &block)
	if err != nil {
		log.Printf("Failed to unmarshal block: %v", err)
		return
	}

	err = p.Blockchain.AddBlock(&block, p.NodeID)
	if err != nil {
		log.Printf("Failed to add block from peer: %v", err)
		return
	}

	log.Printf("Successfully added block from peer")
	p.BroadcastMessage(string(blockData))
}

// BroadcastTransaction broadcasts a transaction to all connected peers.
func (p *P2PNetwork) BroadcastTransaction(tx *blockchain.Transaction) {
	txData, err := json.Marshal(tx)
	if err != nil {
		log.Printf("Failed to marshal transaction: %v", err)
		return
	}
	p.BroadcastMessage(string(txData))
}

// BroadcastBlock broadcasts a block to all connected peers.
func (p *P2PNetwork) BroadcastBlock(block *blockchain.MainBlock) {
	blockData, err := json.Marshal(block)
	if err != nil {
		log.Printf("Failed to marshal block: %v", err)
		return
	}
	p.BroadcastMessage(string(blockData))
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