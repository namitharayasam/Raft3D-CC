package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// RaftServer handles HTTP requests
type RaftServer struct {
	node *RaftNode
}

// VoteRequestHandler handles incoming vote requests
func (rs *RaftServer) VoteRequestHandler(w http.ResponseWriter, r *http.Request) {
	var args RequestVoteArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	reply := &RequestVoteReply{}
	rs.node.RequestVote(&args, reply)

	json.NewEncoder(w).Encode(reply)
}

// StartElectionHandler manually triggers an election
func (rs *RaftServer) StartElectionHandler(w http.ResponseWriter, r *http.Request) {
	rs.node.StartElection()
	fmt.Fprintf(w, "Election started on node %d", rs.node.id)
}

// AppendEntriesHandler handles log replication requests
func (rs *RaftServer) AppendEntriesHandler(w http.ResponseWriter, r *http.Request) {
	var args AppendEntriesArgs
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	reply := &AppendEntriesReply{}
	rs.node.AppendEntries(&args, reply)

	json.NewEncoder(w).Encode(reply)
}

// StartServer starts an HTTP server for Raft interactions
func StartServer(node *RaftNode) {
	rs := &RaftServer{node: node}
	r := mux.NewRouter()

	// ✅ Root route to confirm the server is running
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is running!"))
	})

	// ✅ Existing routes for Raft API
	r.HandleFunc("/vote", rs.VoteRequestHandler).Methods("POST")
	r.HandleFunc("/start-election", rs.StartElectionHandler).Methods("GET")

	// ✅ NEW: Route for AppendEntries (Log Replication)
	r.HandleFunc("/append-entries", rs.AppendEntriesHandler).Methods("POST")

	log.Println("Starting HTTP server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
