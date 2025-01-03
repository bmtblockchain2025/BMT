package network

import (
	"BMT-Blockchain/src/blockchain"
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

// Node represents a single node in the P2P network.
type Node struct {
	ID       string                // Unique ID for the node
	Address  string                // Network address (e.g., "127.0.0.1:8080")
	Peers    map[string]*Peer      // List of connected peers
	Blockchain *blockchain.Blockchain // Blockchain managed by this node
	mutex    sync.Mutex            // Mutex for thread safety
}

// Peer represents a connected peer in the network.
type Peer struct {
	ID      string // Unique ID of the peer
	Address string // Network address of the peer
}

// NewNode creates a new node with a given ID and address.
func NewNode(id, address string) *Node {
	return &Node{
		ID:         id,
		Address:    address,
		Peers:      make(map[string]*Peer),
		Blockchain: blockchain.NewBlockchain(),
	}
}

// Start launches the node to listen for incoming connections.
func (node *Node) Start() {
	fmt.Printf("Node %s is starting at %s...\n", node.ID, node.Address)
	listener, err := net.Listen("tcp", node.Address)
	if err != nil {
		fmt.Printf("Error starting node: %v\n", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go node.handleConnection(conn)
	}
}

// handleConnection handles incoming messages from peers.
func (node *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read incoming data
	decoder := json.NewDecoder(conn)
	var message map[string]interface{}
	if err := decoder.Decode(&message); err != nil {
		fmt.Printf("Error decoding message: %v\n", err)
		return
	}

	// Handle different message types
	messageType, ok := message["type"].(string)
	if !ok {
		fmt.Println("Invalid message format")
		return
	}

	switch messageType {
	case "new_block":
		node.handleNewBlock(message)
	case "sync_request":
		node.handleSyncRequest(conn)
	default:
		fmt.Printf("Unknown message type: %s\n", messageType)
	}
}

// handleNewBlock processes a new block received from a peer.
func (node *Node) handleNewBlock(message map[string]interface{}) {
	blockData, err := json.Marshal(message["block"])
	if err != nil {
		fmt.Printf("Error parsing block: %v\n", err)
		return
	}

	var newBlock blockchain.Block
	if err := json.Unmarshal(blockData, &newBlock); err != nil {
		fmt.Printf("Error unmarshalling block: %v\n", err)
		return
	}

	node.mutex.Lock()
	defer node.mutex.Unlock()

	// Validate and add the new block
	if node.Blockchain.GetLatestBlock().Hash == newBlock.PreviousHash {
		node.Blockchain.Chain = append(node.Blockchain.Chain, &newBlock)
		fmt.Printf("Block added to node %s: %+v\n", node.ID, newBlock)
	} else {
		fmt.Printf("Invalid block received by node %s\n", node.ID)
	}
}

// handleSyncRequest sends the entire blockchain to the requesting peer.
func (node *Node) handleSyncRequest(conn net.Conn) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(node.Blockchain.Chain); err != nil {
		fmt.Printf("Error sending blockchain: %v\n", err)
	}
}

// Connect adds a peer to the node's list of connected peers.
func (node *Node) Connect(peerID, peerAddress string) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if _, exists := node.Peers[peerID]; !exists {
		node.Peers[peerID] = &Peer{
			ID:      peerID,
			Address: peerAddress,
		}
		fmt.Printf("Node %s connected to peer %s at %s\n", node.ID, peerID, peerAddress)
	}
}
