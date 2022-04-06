package raft

import (
	"sync"
	"time"
)

type Role int

const (
	UnknownRole   Role = 0
	CandidateRole Role = 1
	FollowerRole  Role = 2
	LeaderRole    Role = 3
)

type State struct {
	mu           sync.Mutex
	role         Role
	lastAccessed time.Time // last time the node access the heartbeat from leader

	// persistent state on all servers
	currentTerm int
	voteFor     int
	logEntries  []*LogEntry
	lastLogIdx  int
	// volatile state on all servers
	commitIdx   int // index of the highest log entry known to be committed
	lastApplied int // index of the highest log entry applied to statue machine
	// volatile state on leaders, reinitialized after election
	nextIndex  []int
	matchIndex []int
}

type Node struct {
	state *State

	Id    int
	peers []*Node
	me    int
}

func (n *Node) Lock() {
	n.state.mu.Lock()
}

func (n *Node) Unlock() {
	n.state.mu.Unlock()
}
