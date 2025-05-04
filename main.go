package main

import (
	"flag"
	"sync"
	"fmt"

	"cs4049-assignment2-A-i200438_i200650/kvstore"
	"cs4049-assignment2-A-i200438_i200650/server"
	"cs4049-assignment2-A-i200438_i200650/raft"
)

func main() {
	// Parse command line flags
	id := flag.Int("id", 1, "Node ID")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	// Print node information
	fmt.Printf("Starting node %d on port %d\n", *id, *port)

	// Create the key-value store
	kvStore := kvstore.NewKVStore()

	// Initialize peers - in a real system, this would be configured externally
	peers := []int{2, 3} // Example peers - adjust as needed
	
	// Create KVRaft instance
	kvRaft := raft.NewKVRaft(*id, peers)
	
	// Create and start the server
	serverInstance := server.NewServer(kvStore, kvRaft)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverInstance.StartServer(*port)
	}()

	wg.Wait()
}