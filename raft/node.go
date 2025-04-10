// raft/node.go
package raftnode

import (
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"raft3d/fsm"
)

func NewRaftNode(nodeID, raftDir, bindAddr string) (*raft.Raft, *fsm.FSM, error) {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(nodeID)

	if err := os.MkdirAll(raftDir, 0700); err != nil {
		return nil, nil, err
	}

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-log.bolt"))
	if err != nil {
		return nil, nil, err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(raftDir, "raft-stable.bolt"))
	if err != nil {
		return nil, nil, err
	}

	snapshotStore, err := raft.NewFileSnapshotStore(raftDir, 2, os.Stderr)
	if err != nil {
		return nil, nil, err
	}

	// ✅ Use TCP transport instead of inmem (fixes cross-node comms)
	transport, err := raft.NewTCPTransport(bindAddr, nil, 3, 10*time.Second, os.Stderr)
	if err != nil {
		return nil, nil, err
	}

	f := fsm.NewFSM()

	r, err := raft.NewRaft(config, f, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, nil, err
	}

	// ✅ Bootstrap cluster only on fresh state and if node1
	hasState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		return nil, nil, err
	}

	if !hasState && nodeID == "node1" {
		conf := raft.Configuration{
			Servers: []raft.Server{
				{ID: "node1", Address: raft.ServerAddress(bindAddr)},
			},
		}
		r.BootstrapCluster(conf)
	}

	return r, f, nil
}
