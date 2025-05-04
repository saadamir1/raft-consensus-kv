package raft

import (
	"fmt"
	"sync"
)

func (rn *RaftNode) startElection() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	rn.state = Candidate
	rn.currentTerm++
	rn.votedFor = rn.id
	votesReceived := 1 // Vote for itself

	var wg sync.WaitGroup
	mu := sync.Mutex{}
	for _, peer := range rn.peers {
		wg.Add(1)
		go func(peerID int) {
			defer wg.Done()
			voteGranted := rn.requestVote(peerID)
			if voteGranted {
				mu.Lock()
				votesReceived++
				mu.Unlock()
			}
		}(peer)
	}
	wg.Wait()

	if votesReceived > len(rn.peers)/2 {
		rn.becomeLeader()
	} else {
		rn.resetElectionTimer()
	}
}

func (rn *RaftNode) requestVote(peerID int) bool {
	args := RequestVoteArgs{
		Term:         rn.currentTerm,
		CandidateID:  rn.id,
		LastLogIndex: len(rn.log) - 1,
	}
	reply := RequestVoteReply{}
	ok := call(peerID, "RaftNode.RequestVote", args, &reply)
	if ok && reply.VoteGranted {
		return true
	}
	return false
}

func (rn *RaftNode) becomeLeader() {
	rn.state = Leader
	rn.leaderID = rn.id
	fmt.Printf("Node %d became leader in term %d\n", rn.id, rn.currentTerm)

	for _, peer := range rn.peers {
		rn.nextIndex[peer] = len(rn.log)
		rn.matchIndex[peer] = 0
	}

	go rn.sendHeartbeats()
}

// RequestVote RPC handler
func (rn *RaftNode) RequestVote(args RequestVoteArgs, reply *RequestVoteReply) error {
	rn.mu.Lock()
	defer rn.mu.Unlock()
	
	reply.Term = rn.currentTerm
	reply.VoteGranted = false
	
	// 1. Reply false if term < currentTerm
	if args.Term < rn.currentTerm {
		return nil
	}
	
	// If RPC term is higher than our term, update term and step down
	if args.Term > rn.currentTerm {
		rn.currentTerm = args.Term
		rn.state = Follower
		rn.votedFor = -1
	}
	
	// 2. If votedFor is null or candidateId, and candidate's log is at least as up-to-date as receiver's log, grant vote
	if (rn.votedFor == -1 || rn.votedFor == args.CandidateID) &&
	   args.LastLogIndex >= len(rn.log)-1 {
		reply.VoteGranted = true
		rn.votedFor = args.CandidateID
		rn.resetElectionTimer() // Reset election timer when granting vote
	}
	
	return nil
}