// main.go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/hashicorp/raft"
    "raft3d/fsm"
    "raft3d/models"
    "raft3d/raft"
)

var raftNode *raft.Raft
var fsmInstance *fsm.FSM

func main() {
    // ✅ Read args: nodeID, raftBindAddr, httpPort
    if len(os.Args) != 4 {
        log.Fatalf("Usage: go run main.go <nodeID> <raftBindAddr> <httpPort>")
    }

    nodeID := os.Args[1]
    raftBindAddr := os.Args[2]
    httpPort := os.Args[3]
    raftDir := "raft-data-" + nodeID // separate storage per node

    // ✅ Start Raft node
    var err error
    raftNode, fsmInstance, err = raftnode.NewRaftNode(nodeID, raftDir, raftBindAddr)
    if err != nil {
        log.Fatalf("Failed to start Raft node: %v", err)
        os.Exit(1)
    }

    // ✅ Home check
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Raft3D node [%s] is running!", nodeID)
    })

    // ✅ /api/v1/printers — Add or List printers
    http.HandleFunc("/api/v1/printers", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodPost:
            if raftNode.State() != raft.Leader {
                http.Error(w, "This node is not the leader", http.StatusBadRequest)
                return
            }

            var p models.Printer
            if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
                http.Error(w, "Invalid JSON", http.StatusBadRequest)
                return
            }

            data, err := json.Marshal(p)
            if err != nil {
                http.Error(w, "Marshal failed", http.StatusInternalServerError)
                return
            }

            cmd := append([]byte("printer:"), data...)
            future := raftNode.Apply(cmd, 5000)
            if err := future.Error(); err != nil {
                http.Error(w, "Raft apply failed: "+err.Error(), http.StatusInternalServerError)
                return
            }

            fmt.Fprintf(w, "Printer %s added\n", p.ID)

        case http.MethodGet:
            printers := fsmInstance.GetAllPrinters()
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(printers)

        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    // ✅ /join — Accept a new node into the cluster (called on leader)
    http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
        if raftNode.State() != raft.Leader {
            http.Error(w, "Only leader can process join requests", http.StatusForbidden)
            return
        }

        var req struct {
            ID      string `json:"id"`
            Address string `json:"address"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        future := raftNode.AddVoter(raft.ServerID(req.ID), raft.ServerAddress(req.Address), 0, 0)
        if err := future.Error(); err != nil {
            http.Error(w, "Failed to add voter: "+err.Error(), http.StatusInternalServerError)
            return
        }

        fmt.Fprintf(w, "Node %s added at %s\n", req.ID, req.Address)
    })

    // ✅ /status — Show current Raft state (Leader/Follower/Candidate)
    http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
        status := map[string]string{
            "node":  nodeID,
            "state": raftNode.State().String(),
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(status)
    })

    // ✅ Start HTTP server
    fmt.Printf("HTTP server started at http://localhost:%s\n", httpPort)
    log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
