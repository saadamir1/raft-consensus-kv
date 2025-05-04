package raft

import (
	"fmt"
	"net"
	"net/rpc"
	"log"
)

// StartRPCServer starts the RPC server for a RaftNode
func (rn *RaftNode) StartRPCServer() {
	// Register the RPC handlers
	err := rpc.Register(rn)
	if err != nil {
		log.Fatalf("Error registering RPC server: %v", err)
	}
	
	// Set up the listener
	addr := fmt.Sprintf(":%d", 9000+rn.id)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error listening on port %s: %v", addr, err)
	}
	
	// Serve RPC
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Accept error: %v", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
	
	log.Printf("RPC server started on port %d", 9000+rn.id)
}

func call(peerID int, method string, args interface{}, reply interface{}) bool {
	client, err := rpc.Dial("tcp", fmt.Sprintf("localhost:%d", 9000+peerID))
	if err != nil {
		return false
	}
	defer client.Close()

	err = client.Call(method, args, reply)
	return err == nil
}

type RequestVoteArgs struct {
	Term         int
	CandidateID  int
	LastLogIndex int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}