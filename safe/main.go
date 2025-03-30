package main

func main() {
	// Create Raft node with ID 1 and two peers as string addresses
	peers := []string{"localhost:5002", "localhost:5003"} // Use proper addresses as strings
	node := NewRaftNode(1, peers)

	// Start HTTP server to expose APIs
	StartServer(node)
}
