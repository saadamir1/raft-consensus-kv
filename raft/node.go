package raft

type Node struct {
	ID    int
	State RaftState
}