package raft

type State struct {
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
