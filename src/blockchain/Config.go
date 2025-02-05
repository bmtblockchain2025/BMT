package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config defines the configuration for the BMT Blockchain.
type Config struct {
	NodeID         string  `json:"node_id"`
	APIPort        string  `json:"api_port"`
	P2PAddress     string  `json:"p2p_address"`
	MaxBlockSize   int     `json:"max_block_size"`
	BlockTime      int     `json:"block_time"`
	ConsensusType  string  `json:"consensus_type"`
	PoSWeight      float64 `json:"pos_weight"`
	PoHTimeFactor  float64 `json:"poh_time_factor"`
	BFTThreshold   float64 `json:"bft_threshold"`
	DPoSValidators int     `json:"dpos_validators"`
	TransactionFee float64 `json:"transaction_fee"`
	StakingReward  float64 `json:"staking_reward"`
	MinStakeAmount float64 `json:"min_stake_amount"`
	DatabasePath   string  `json:"database_path"`
}

// DefaultConfig provides default configuration values.
var DefaultConfig = Config{
	NodeID:         "bmt-node-1",
	APIPort:        "8080",
	P2PAddress:     "0.0.0.0:9000",
	MaxBlockSize:   10,    // Max block size in MB
	BlockTime:      10,    // Block creation time in seconds
	ConsensusType:  "Hybrid", // PoS + PoH + BFT + DPoS
	PoSWeight:      0.5,   // Weight of PoS in hybrid consensus
	PoHTimeFactor:  0.2,   // Influence of historical time-based validation
	BFTThreshold:   0.67,  // Byzantine Fault Tolerance agreement threshold
	DPoSValidators: 21,    // Number of validators in DPoS
	TransactionFee: 0.01,  // Transaction fee in BMT
	StakingReward:  2.0,   // Reward per block
	MinStakeAmount: 50.0,  // Minimum stake required to be a validator
	DatabasePath:   "data/bmt_blockchain.db",
}

// LoadConfig loads the configuration from a JSON file or returns default values.
func LoadConfig() Config {
	file, err := os.Open("config.json")
	if err != nil {
		log.Println("Config file not found, using default settings.")
		return DefaultConfig
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	err = decoder.Decode(&cfg)
	if err != nil {
		log.Println("Error decoding config file, using default settings.")
		return DefaultConfig
	}

	log.Println("Configuration loaded successfully.")
	return cfg
}
