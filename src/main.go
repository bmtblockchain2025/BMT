package main

import (
	"src/blockchain"
	"src/network"
	"src/api"
	"src/config"
	"log"
	"sync"
)

func main() {
	log.Println("Starting BMT Blockchain Node...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize blockchain
	bc := blockchain.NewBlockchain()
	p2pNetwork := network.NewP2PNetwork(cfg.NodeID, cfg.P2PAddress, nil, bc)

	// Initialize API
	apiServer := api.NewAPI(nil, nil, bc)

	var wg sync.WaitGroup
	wg.Add(2)

	// Start P2P Network
	go func() {
		defer wg.Done()
		log.Println("Starting P2P Network on", cfg.P2PAddress)
		p2pNetwork.Start()
	}()

	// Start API Server
	go func() {
		defer wg.Done()
		log.Println("Starting API Server on port", cfg.APIPort)
		apiServer.StartAPI(cfg.APIPort)
	}()

	wg.Wait()
}
