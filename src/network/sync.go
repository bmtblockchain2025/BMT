package network

import (
	"BMT-Blockchain/src/blockchain"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// Synchronizer handles blockchain synchronization across nodes.
type Synchronizer struct {
	mutex sync.Mutex
}

// RequestBlockchain requests the blockchain data from a peer.
func (s *Synchronizer) RequestBlockchain(peerAddress string) ([]*blockchain.Block, error) {
	conn, err := net.Dial("tcp", peerAddress)
	if err != nil {
		return nil, fmt.Errorf("error connecting to peer: %v", err)
	}
	defer conn.Close()

	// Send sync request message
	message := map[string]string{"type": "sync_request"}
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(message); err != nil {
		return nil, fmt.Errorf("error sending sync request: %v", err)
	}

	// Receive blockchain data
	var blockchainData []*blockchain.Block
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&blockchainData); err != nil {
		return nil, fmt.Errorf("error decoding blockchain data: %v", err)
	}

	return blockchainData, nil
}

// UpdateBlockchain replaces the current blockchain with the received blockchain if valid.
func (s *Synchronizer) UpdateBlockchain(node *Node, newBlockchain []*blockchain.Block) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(newBlockchain) <= len(node.Blockchain.Chain) {
		return fmt.Errorf("received blockchain is not longer than the current blockchain")
	}

	// Validate the received blockchain
	tempBlockchain := &blockchain.Blockchain{Chain: newBlockchain}
	if !tempBlockchain.IsValid() {
		return fmt.Errorf("received blockchain is invalid")
	}

	// Replace the current blockchain
	node.Blockchain.Chain = newBlockchain
	fmt.Printf("Node %s updated its blockchain with new data.\n", node.ID)
	return nil
}
