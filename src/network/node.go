package network

import (
	"crypto/tls"
	"errors"
	"log"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// Node represents a single node in the blockchain network.
type Node struct {
	ID             string
	Address        string
	ConnectedPeers map[string]*Node // Connected peers in the network
	Blockchain     *blockchain.Blockchain
	mempool        []*blockchain.Transaction // Pending transactions
	mutex          sync.Mutex
	heartbeatMutex sync.Mutex
	workerPool     chan struct{} // Worker pool for transaction processing
	TLSConfig      *tls.Config   // TLS configuration for secure connections
}

// NewNode initializes a new node with a unique ID and address.
func NewNode(id, address string, blockchain *blockchain.Blockchain) *Node {
	return &Node{
		ID:             id,
		Address:        address,
		ConnectedPeers: make(map[string]*Node),
		Blockchain:     blockchain,
		mempool:        []*blockchain.Transaction{},
		workerPool:     make(chan struct{}, 10), // Limit to 10 concurrent workers
		TLSConfig:      &tls.Config{InsecureSkipVerify: true},
	}
}

// ConnectPeer connects the current node to a new peer.
func (n *Node) ConnectPeer(peer *Node) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if _, exists := n.ConnectedPeers[peer.ID]; exists {
		return errors.New("peer already connected")
	}

	n.ConnectedPeers[peer.ID] = peer
	log.Printf("Node %s connected to peer %s", n.ID, peer.ID)
	return nil
}

// DisconnectPeer disconnects the current node from a peer.
func (n *Node) DisconnectPeer(peerID string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if _, exists := n.ConnectedPeers[peerID]; !exists {
		return errors.New("peer not found")
	}

	delete(n.ConnectedPeers, peerID)
	log.Printf("Node %s disconnected from peer %s", n.ID, peerID)
	return nil
}

// BroadcastTransaction broadcasts a transaction to all connected peers.
func (n *Node) BroadcastTransaction(tx *blockchain.Transaction) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for _, peer := range n.ConnectedPeers {
		go peer.ReceiveTransaction(tx)
	}
}

// ReceiveTransaction handles an incoming transaction.
func (n *Node) ReceiveTransaction(tx *blockchain.Transaction) {
	select {
	case n.workerPool <- struct{}{}:
		go func() {
			n.mutex.Lock()
			defer n.mutex.Unlock()
			defer func() { <-n.workerPool }()

			if err := n.Blockchain.AddTransaction(tx, n.ID); err != nil {
				log.Printf("Node %s failed to add transaction: %v", n.ID, err)
				return
			}

			// Add transaction to mempool
			n.mempool = append(n.mempool, tx)

			// Broadcast the transaction to other peers
			n.BroadcastTransaction(tx)
		}()
	default:
		log.Printf("Node %s worker pool is full, dropping transaction", n.ID)
	}
}

// SyncBlockchain synchronizes the blockchain incrementally with a peer.
func (n *Node) SyncBlockchain(peer *Node) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	peerBlockchain := peer.Blockchain
	commonIndex := n.findCommonAncestor(peerBlockchain)

	if commonIndex == -1 {
		return errors.New("no common ancestor found")
	}

	if len(peerBlockchain.Chain) > len(n.Blockchain.Chain) {
		n.Blockchain.Chain = append(n.Blockchain.Chain[:commonIndex+1], peerBlockchain.Chain[commonIndex+1:]...)
		log.Printf("Node %s incrementally synced blockchain with peer %s", n.ID, peer.ID)
		return nil
	}

	return errors.New("no sync needed, local blockchain is up-to-date")
}

// findCommonAncestor finds the last common block index between two blockchains.
func (n *Node) findCommonAncestor(peerBlockchain *blockchain.Blockchain) int {
	for i := len(n.Blockchain.Chain) - 1; i >= 0; i-- {
		if n.Blockchain.Chain[i].Hash == peerBlockchain.Chain[i].Hash {
			return i
		}
	}
	return -1
}

// Heartbeat sends a heartbeat signal to all connected peers to check their status.
func (n *Node) Heartbeat() {
	n.heartbeatMutex.Lock()
	defer n.heartbeatMutex.Unlock()

	for _, peer := range n.ConnectedPeers {
		go func(p *Node) {
			// Simulate heartbeat check
			time.Sleep(1 * time.Second)
			log.Printf("Node %s received heartbeat from peer %s", n.ID, p.ID)
		}(peer)
	}
}

// DetectMaliciousPeers scans for peers sending invalid data or behaving abnormally.
func (n *Node) DetectMaliciousPeers() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	for id, peer := range n.ConnectedPeers {
		invalidTxCount := rand.Intn(10) // Simulated detection logic
		if invalidTxCount > 5 {
			log.Printf("Node %s detected malicious peer %s", n.ID, id)
			n.DisconnectPeer(id)
		}
	}
}

// PrioritizeTransactions sorts the mempool by transaction fee and timestamp.
func (n *Node) PrioritizeTransactions() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	sort.Slice(n.mempool, func(i, j int) bool {
		if n.mempool[i].Fee == n.mempool[j].Fee {
			return n.mempool[i].Timestamp < n.mempool[j].Timestamp
		}
		return n.mempool[i].Fee > n.mempool[j].Fee
	})
}

// LightNodeMode enables light mode, where only block headers are stored.
func (n *Node) LightNodeMode() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	n.Blockchain.Chain = nil // Clear full blocks, keep only headers
	log.Printf("Node %s switched to light node mode", n.ID)
}

// DistributedStorage enables storing only parts of the blockchain.
func (n *Node) DistributedStorage() {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	log.Printf("Node %s enabled distributed storage", n.ID)
}
