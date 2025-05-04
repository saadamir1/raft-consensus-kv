package raft

import (
	"encoding/json"
	"sync"
	"time"
)

// Operation types
const (
	PUT    = "PUT"
	APPEND = "APPEND"
	GET    = "GET"
)

// Command represents a client command to be executed
type Command struct {
	Op    string // Operation type: PUT, APPEND, GET
	Key   string
	Value string
}

// KVRaft connects the KV store with the Raft consensus
type KVRaft struct {
	mu      sync.Mutex
	store   map[string]string
	node    *RaftNode
	applyCh chan ApplyMsg
}

// NewKVRaft creates a new KVRaft instance
func NewKVRaft(id int, peers []int) *KVRaft {
	kv := &KVRaft{
		store:   make(map[string]string),
		applyCh: make(chan ApplyMsg, 100), // Buffered channel to prevent blocking
	}
	kv.node = NewRaftNode(id, peers, kv.applyCh)

	// Start the apply loop
	go kv.applyCommands()

	// Return the initialized instance
	return kv
}

// applyCommands applies committed log entries to the state machine
func (kv *KVRaft) applyCommands() {
	for msg := range kv.applyCh {
		if !msg.CommandValid {
			continue
		}

		// Lock the store for modification
		kv.mu.Lock()
		cmd := msg.Command

		switch cmd.Op {
		case PUT:
			kv.store[cmd.Key] = cmd.Value
		case APPEND:
			kv.store[cmd.Key] += cmd.Value
		case GET:
			// GET doesn't modify state
		}
		kv.mu.Unlock()
	}
}

// Put adds a key-value pair to the store via Raft consensus
func (kv *KVRaft) Put(key, value string) bool {
	cmd := Command{
		Op:    PUT,
		Key:   key,
		Value: value,
	}
	return kv.submitCommand(cmd)
}

// Append appends a value to an existing key via Raft consensus
func (kv *KVRaft) Append(key, value string) bool {
	cmd := Command{
		Op:    APPEND,
		Key:   key,
		Value: value,
	}
	return kv.submitCommand(cmd)
}

// Get retrieves a value for a key via Raft consensus
func (kv *KVRaft) Get(key string) (string, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, exists := kv.store[key]
	return val, exists
}

// submitCommand submits a command to the Raft cluster
func (kv *KVRaft) submitCommand(cmd Command) bool {
	// Convert command to JSON
	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return false
	}

	// Submit to Raft for consensus
	success := kv.node.ProposeCommand(string(cmdBytes))
	
	// Wait a bit for command to be applied
	time.Sleep(100 * time.Millisecond)
	
	return success
}