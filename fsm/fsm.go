// fsm/fsm.go
package fsm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	"raft3d/models"
)

// FSM implements the raft.FSM interface for storing printers
type FSM struct {
	mu       sync.Mutex
	printers map[string]models.Printer
}

// NewFSM creates a new FSM instance
func NewFSM() *FSM {
	return &FSM{
		printers: make(map[string]models.Printer),
	}
}

// Apply is called once a log entry is committed
func (f *FSM) Apply(logEntry *raft.Log) interface{} {
	if bytes.HasPrefix(logEntry.Data, []byte("printer:")) {
		var p models.Printer
		err := json.Unmarshal(logEntry.Data[len("printer:"):], &p)
		if err != nil {
			fmt.Println("[FSM] Failed to unmarshal printer:", err)
			return nil
		}

		f.mu.Lock()
		f.printers[p.ID] = p
		f.mu.Unlock()
		fmt.Printf("[FSM] Applied printer: %s\n", p.ID)
	}
	return nil
}

// Snapshot returns a snapshot of the FSM state
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &fsmSnapshot{printers: f.printers}, nil
}

// Restore recovers the FSM from a snapshot
func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	// No restore logic yet
	return nil
}

// fsmSnapshot implements raft.FSMSnapshot
type fsmSnapshot struct {
	printers map[string]models.Printer
}

// Persist writes the snapshot to the sink
func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	sink.Cancel() // No snapshot persistence implemented
	return nil
}

// Release is called when done with snapshot
func (s *fsmSnapshot) Release() {}

// GetAllPrinters returns all printers from FSM
func (f *FSM) GetAllPrinters() []models.Printer {
	f.mu.Lock()
	defer f.mu.Unlock()

	var result []models.Printer
	for _, p := range f.printers {
		result = append(result, p)
	}
	return result
}
