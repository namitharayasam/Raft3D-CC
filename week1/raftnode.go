package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// RaftNode represents a node in the Raft cluster
type RaftNode struct {
	mu              sync.Mutex
	id              int
	state           string // "Follower", "Candidate", "Leader"
	term            int
	votedFor        int
	peers           []int // Other node IDs
	electionTimeout time.Duration
}

// RequestVoteArgs is sent to request votes
type RequestVoteArgs struct {
	Term        int
	CandidateID int
}

// RequestVoteReply is the response to a vote request
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// RequestVote handles incoming vote requests
func (rn *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term > rn.term {
		rn.term = args.Term
		rn.votedFor = args.CandidateID
		reply.Term = rn.term
		reply.VoteGranted = true
	} else {
		reply.Term = rn.term
		reply.VoteGranted = false
	}

	return nil
}

// StartElection triggers an election process
func (rn *RaftNode) StartElection() {
	rn.mu.Lock()
	rn.state = "Candidate"
	rn.term++
	rn.votedFor = rn.id
	rn.mu.Unlock()

	fmt.Printf("Node %d starting election for term %d\n", rn.id, rn.term)

	// Simulated vote request (normally, you'd send RPCs here)
	votes := 1
	for _, peer := range rn.peers {
		fmt.Printf("Node %d requesting vote from Node %d\n", rn.id, peer)
		votes++ // Simulating a granted vote
	}

	if votes > len(rn.peers)/2 {
		rn.mu.Lock()
		rn.state = "Leader"
		rn.mu.Unlock()
		fmt.Printf("Node %d became the Leader for term %d\n", rn.id, rn.term)
	}
}

// NewRaftNode creates a new node
func NewRaftNode(id int, peers []int) *RaftNode {
	return &RaftNode{
		id:              id,
		state:           "Follower",
		peers:           peers,
		term:            0,
		votedFor:        -1,
		electionTimeout: time.Duration(rand.Intn(300-150)+150) * time.Millisecond,
	}
}

