package main

func main() {
	// Create Raft node with ID 1 and two peers (2, 3)
	node := NewRaftNode(1, []int{2, 3})

	// Start HTTP server to expose APIs
	StartServer(node)
}

