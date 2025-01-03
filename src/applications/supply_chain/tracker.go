package supply_chain

import (
	"errors"
	"sync"
	"time"
)

// Goods represents an individual item in the supply chain.
type Goods struct {
	ID          string        // Unique identifier for the goods
	Name        string        // Name of the goods
	Owner       string        // Current owner of the goods
	Status      string        // Current status (e.g., "In Transit", "Delivered")
	History     []Transaction // History of transactions related to the goods
	mutex       sync.Mutex    // Mutex for thread safety
}

// Transaction represents a single transaction in the goods' history.
type Transaction struct {
	Timestamp int64  // Timestamp of the transaction
	From      string // Previous owner
	To        string // New owner
	Status    string // Status change (if any)
}

// SupplyChainTracker manages goods and their transactions.
type SupplyChainTracker struct {
	Goods map[string]*Goods // Mapping from goods ID to Goods object
	mutex sync.Mutex        // Mutex for thread safety
}

// NewSupplyChainTracker initializes a new supply chain tracker.
func NewSupplyChainTracker() *SupplyChainTracker {
	return &SupplyChainTracker{
		Goods: make(map[string]*Goods),
	}
}

// AddGoods adds new goods to the supply chain.
func (tracker *SupplyChainTracker) AddGoods(id, name, owner, status string) (*Goods, error) {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	if _, exists := tracker.Goods[id]; exists {
		return nil, errors.New("goods with this ID already exists")
	}

	goods := &Goods{
		ID:     id,
		Name:   name,
		Owner:  owner,
		Status: status,
		History: []Transaction{
			{
				Timestamp: time.Now().Unix(),
				From:      "Manufacturer",
				To:        owner,
				Status:    status,
			},
		},
	}
	tracker.Goods[id] = goods
	return goods, nil
}

// UpdateGoods updates the status or ownership of goods in the supply chain.
func (tracker *SupplyChainTracker) UpdateGoods(id, newOwner, newStatus string) error {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	goods, exists := tracker.Goods[id]
	if !exists {
		return errors.New("goods not found")
	}

	goods.mutex.Lock()
	defer goods.mutex.Unlock()

	goods.Owner = newOwner
	goods.Status = newStatus
	goods.History = append(goods.History, Transaction{
		Timestamp: time.Now().Unix(),
		From:      goods.Owner,
		To:        newOwner,
		Status:    newStatus,
	})
	return nil
}

// GetGoods retrieves the current state and history of goods by ID.
func (tracker *SupplyChainTracker) GetGoods(id string) (*Goods, error) {
	tracker.mutex.Lock()
	defer tracker.mutex.Unlock()

	goods, exists := tracker.Goods[id]
	if !exists {
		return nil, errors.New("goods not found")
	}

	return goods, nil
}
