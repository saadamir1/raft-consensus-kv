package raft

import (
	"encoding/json"
	"time"
)

type LogEntry struct {
	Term    int
	Command string
}

type AppendEntryArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntryReply struct {
	Term    int
	Success bool
}

func (rn *RaftNode) AppendEntries(args AppendEntryArgs, reply *AppendEntryReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	reply.Term = rn.currentTerm
	reply.Success = false
	
	// 1. Reply false if term < currentTerm
	if args.Term < rn.currentTerm {
		return nil
	}
	
	// Reset election timer since we heard from leader
	rn.resetElectionTimer()
	
	// If RPC term is higher than our term, update term and step down
	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.state = Follower
		rn.votedFor = -1
	}
	
	// Update leader ID
	rn.leaderID = args.LeaderID
	
	// 2. Reply false if log doesn't contain an entry at prevLogIndex whose term matches prevLogTerm
	if args.PrevLogIndex >= len(rn.log) || 
	   (args.PrevLogIndex >= 0 && rn.log[args.PrevLogIndex].Term != args.PrevLogTerm) {
		return nil
	}
	
	// 3. If an existing entry conflicts with a new one, delete the existing entry and all that follow it
	newEntryIndex := args.PrevLogIndex + 1
	for i, entry := range args.Entries {
		if newEntryIndex+i < len(rn.log) {
			if rn.log[newEntryIndex+i].Term != entry.Term {
				// Cut the log here
				rn.log = rn.log[:newEntryIndex+i]
				break
			}
		} else {
			break
		}
	}
	
	// 4. Append any new entries not already in the log
	startIdx := 0
	if newEntryIndex < len(rn.log) {
		startIdx = len(rn.log) - newEntryIndex
	}
	
	if startIdx < len(args.Entries) {
		rn.log = append(rn.log, args.Entries[startIdx:]...)
	}
	
	// 5. If leaderCommit > commitIndex, set commitIndex = min(leaderCommit, index of last new entry)
	if args.LeaderCommit > rn.commitIndex {
		rn.commitIndex = min(args.LeaderCommit, len(rn.log)-1)
		go rn.applyCommittedEntries()
	}
	
	reply.Success = true
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (rn *RaftNode) sendHeartbeats() {
	for rn.state == Leader {
		for _, peer := range rn.peers {
			go rn.sendAppendEntries(peer)
		}
		time.Sleep(rn.heartbeatInterval)
	}
}

func (rn *RaftNode) sendAppendEntries(peerID int) {
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
		return
	}
	
	if reply.Success {
		rn.mu.Lock()
		rn.nextIndex[peerID] = prevLogIndex + len(entries) + 1
		rn.matchIndex[peerID] = prevLogIndex + len(entries)
		rn.mu.Unlock()
	} else {
		// Log inconsistency, decrement nextIndex and retry
		rn.mu.Lock()
		if rn.nextIndex[peerID] > 1 {
			rn.nextIndex[peerID]--
		}
		rn.mu.Unlock()
	}
}

// applyCommittedEntries applies committed log entries to the state machine
func (rn *RaftNode) applyCommittedEntries() {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	for i := rn.lastApplied + 1; i <= rn.commitIndex; i++ {
		if i < len(rn.log) {
			var cmd Command
			err := json.Unmarshal([]byte(rn.log[i].Command), &cmd)
			if err == nil {
				msg := ApplyMsg{
					CommandValid: true,
					Command:      cmd,
					CommandIndex: i,
				}
				rn.applyCh <- msg
			}
			rn.lastApplied = i
		}
	}
}