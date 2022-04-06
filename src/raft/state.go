package raft

type Role int

const (
	UnknownRole   Role = 0
	CandidateRole Role = 1
	FollowerRole  Role = 2
	LeaderRole    Role = 3
)

type State struct {
	role Role
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
