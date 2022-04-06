package raft

type LogEntry struct {
	Term  int64
	Index int64
}
