package raft

import (
	"math/rand"
	"sync"
	"time"
)

type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

type RaftNode struct {
	mu                sync.Mutex
	id                int
	state             RaftState
	peers             []int
	currentTerm       int
	votedFor          int
	log               []LogEntry
	commitIndex       int
	lastApplied       int
	nextIndex         map[int]int
	matchIndex        map[int]int
	leaderID          int
	electionTimer     *time.Timer
	heartbeatInterval time.Duration
	applyCh           chan ApplyMsg
}

// ApplyMsg is sent by Raft to notify that a log entry is committed
type ApplyMsg struct {
	CommandValid bool
	Command      Command
	CommandIndex int
}

func NewRaftNode(id int, peers []int, applyCh chan ApplyMsg) *RaftNode {
	node := &RaftNode{
		id:                id,
		state:             Follower,
		peers:             peers,
		currentTerm:       0,
		votedFor:          -1,
		log:               []LogEntry{},
		commitIndex:       0,
		lastApplied:       0,
		nextIndex:         make(map[int]int),
		matchIndex:        make(map[int]int),
		leaderID:          -1,
		heartbeatInterval: 50 * time.Millisecond,
		applyCh:           applyCh,
	}
	
	// Initialize with a dummy entry
	node.log = append(node.log, LogEntry{Term: 0})
	
	node.resetElectionTimer()
	
	// Start the RPC server
	node.StartRPCServer()
	
	return node
}

func (rn *RaftNode) resetElectionTimer() {
	if rn.electionTimer != nil {
		rn.electionTimer.Stop()
	}
	rn.electionTimer = time.AfterFunc(randomElectionTimeout(), func() {
		rn.startElection()
	})
}

func randomElectionTimeout() time.Duration {
	return time.Millisecond * time.Duration(150+rand.Intn(150))
}

// ProposeCommand adds a command to the log and replicates it
func (rn *RaftNode) ProposeCommand(command string) bool {
	rn.mu.Lock()
	
	// Only the leader can propose commands
	if rn.state != Leader {
		rn.mu.Unlock()
		return false
	}
	
	// Create a new log entry
	entry := LogEntry{
		Term:    rn.currentTerm,
		Command: command,
	}
	
	// Add to leader's log
	rn.log = append(rn.log, entry)
	logIndex := len(rn.log) - 1
	
	rn.mu.Unlock()
	
	// Try to replicate to majority
	replicatedCount := 1 // Count the leader
	var replicatedMu sync.Mutex
	
	var wg sync.WaitGroup
	for _, peer := range rn.peers {
		wg.Add(1)
		go func(peerID int) {
			defer wg.Done()
			if rn.replicateLogTo(peerID, logIndex) {
				replicatedMu.Lock()
				replicatedCount++
				replicatedMu.Unlock()
			}
		}(peer)
	}
	wg.Wait()
	
	// Check if majority have replicated
	if replicatedCount > (len(rn.peers)+1)/2 {
		rn.mu.Lock()
		if rn.commitIndex < logIndex {
			rn.commitIndex = logIndex
			go rn.applyCommittedEntries()
		}
		rn.mu.Unlock()
		return true
	}
	
	return false
}

// replicateLogTo replicates log entries to a specific peer
func (rn *RaftNode) replicateLogTo(peerID int, upToIndex int) bool {
	rn.mu.Lock()
	prevLogIndex := rn.nextIndex[peerID] - 1
	prevLogTerm := 0
	if prevLogIndex >= 0 && prevLogIndex < len(rn.log) {
		prevLogTerm = rn.log[prevLogIndex].Term
	}
	
	entries := rn.log[rn.nextIndex[peerID]:]
	args := AppendEntryArgs{
		Term:         rn.currentTerm,
		LeaderID:     rn.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: rn.commitIndex,
	}
	rn.mu.Unlock()
	
	reply := AppendEntryReply{}
	ok := call(peerID, "RaftNode.AppendEntries", args, &reply)
	
	if !ok {
		return false
	}
	
	if reply.Success {
		rn.mu.Lock()
		rn.nextIndex[peerID] = upToIndex + 1
		rn.matchIndex[peerID] = upToIndex
		rn.mu.Unlock()
		return true
	} else {
		// Log inconsistency, decrement nextIndex and retry
		rn.mu.Lock()
		if rn.nextIndex[peerID] > 1 {
			rn.nextIndex[peerID]--
		}
		rn.mu.Unlock()
		return rn.replicateLogTo(peerID, upToIndex)
	}
}