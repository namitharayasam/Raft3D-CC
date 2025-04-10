package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

// LogEntry represents a single command in Raft's replicated log
type LogEntry struct {
	Term    int
	Command string
}

// RaftNode represents a node in the Raft cluster
type RaftNode struct {
	mu          sync.Mutex
	id          int
	term        int
	logs        []LogEntry
	isLeader    bool
	peers       []string
	votedFor    int
	commitIndex int
	lastApplied int
}

// AppendEntriesArgs represents the arguments for the AppendEntries RPC
type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

// AppendEntriesReply represents the response to AppendEntries RPC
type AppendEntriesReply struct {
	Term    int
	Success bool
}

// RequestVoteArgs represents the request for voting
type RequestVoteArgs struct {
	Term        int
	CandidateID int
	LastLogIndex int
	LastLogTerm  int
}

// RequestVoteReply represents the reply for voting request
type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

// NewRaftNode initializes a new Raft node
func NewRaftNode(id int, peers []string) *RaftNode {
	return &RaftNode{
		id:       id,
		term:     0,
		isLeader: false,
		peers:    peers,
		votedFor: -1,
	}
}

// AppendEntries handles incoming log replication requests
func (rn *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.term {
		reply.Term = rn.term
		reply.Success = false
		return
	}

	// Append new log entries
	rn.logs = append(rn.logs, args.Entries...)
	reply.Term = args.Term
	reply.Success = true
}

// RequestVote handles incoming vote requests
func (rn *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if args.Term < rn.term {
		reply.Term = rn.term
		reply.VoteGranted = false
		return
	}

	// Grant vote if node hasn't voted yet in this term
	if rn.votedFor == -1 || rn.votedFor == args.CandidateID {
		rn.votedFor = args.CandidateID
		reply.VoteGranted = true
	} else {
		reply.VoteGranted = false
	}
	reply.Term = args.Term
}

// SendHeartbeat sends heartbeats to all followers
func (rn *RaftNode) SendHeartbeat() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if !rn.isLeader {
		return
	}

	for _, peer := range rn.peers {
		go func(peer string) {
			args := AppendEntriesArgs{
				Term:     rn.term,
				LeaderID: rn.id,
				Entries:  []LogEntry{}, // Heartbeat contains no new entries
			}
			reply := &AppendEntriesReply{}

			// Serialize `args` into JSON
			data, err := json.Marshal(args)
			if err != nil {
				log.Printf("Error marshaling heartbeat data: %v", err)
				return
			}

			// Send JSON-encoded `args` in the request body
			resp, err := http.Post("http://"+peer+"/append-entries", "application/json", 
				bytes.NewBuffer(data))
			if err != nil {
				log.Printf("Error sending heartbeat to %s: %v", peer, err)
				return
			}
			defer resp.Body.Close()

			json.NewDecoder(resp.Body).Decode(reply)
			if reply.Term > rn.term {
				rn.term = reply.Term
				rn.isLeader = false
			}
		}(peer)
	}
}


// StartElection starts a new election round
func (rn *RaftNode) StartElection() {
	rn.mu.Lock()
	rn.term++
	rn.isLeader = true // Assume leadership if elected
	rn.mu.Unlock()

	log.Printf("Node %d started an election and became leader", rn.id)
	go rn.SendHeartbeat()
}
