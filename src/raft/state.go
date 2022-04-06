package raft

type State struct {
	// persistent state on all servers
	currentTerm int64
	voteFor     int64
	logEntries  []*LogEntry
	// volatile state on all servers
	commitIdx   int64 // index of the highest log entry known to be committed
	lastApplied int64 // index of the highest log entry applied to statue machine
	// volatile state on leaders, reinitialized after election
	nextIndex  []int64
	matchIndex []int64
}
