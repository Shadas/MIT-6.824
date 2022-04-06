package raft

import "math"

type Node struct {
	state *State
}

func (n *Node) RequestVote(candidateTerm, candidateId, lastLogIdx, lastLogTerm int) (term int, voteGranted bool) {
	if candidateTerm < n.state.currentTerm {
		voteGranted = false
		return
	}
	if n.state.voteFor != 0 && n.state.voteFor != candidateId {
		voteGranted = false
		return
	}
	if lastLogIdx >= n.state.commitIdx {
		voteGranted = true
		term = candidateTerm
	}
	return
}

type AppendEntriesReq struct {
	Term         int // leader's term
	NodeId       int // leader's id
	prevLogIdx   int
	prevLogTerm  int
	leaderCommit int
	entries      []*LogEntry
}

type AppendEntriesResp struct {
	Term    int
	Success bool
}

func (n *Node) AppendEntries(req *AppendEntriesReq) (resp *AppendEntriesResp) {
	resp = &AppendEntriesResp{
		Term:    n.state.currentTerm,
		Success: false,
	}
	if req == nil {
		return
	}
	// check term
	if req.Term < n.state.currentTerm {
		return
	}

	// todo : what if there are logs gap?
	if len(n.state.logEntries) < req.prevLogIdx+1 {
		return
	}

	// check prev log idx if needed
	l := n.state.logEntries[req.prevLogIdx]
	if l.Term != req.prevLogTerm {
		return
	}

	// remove conflict logs
	idx := 0
	for ; idx < len(req.entries); idx++ {
		currIdx := req.prevLogIdx + 1 + idx      // scan idx in the node's log entry slice
		if currIdx > len(n.state.logEntries)-1 { // no conflict
			break
		}
		if n.state.logEntries[currIdx].Term != req.entries[idx].Term { // conflict
			n.state.logEntries = n.state.logEntries[:currIdx]
			n.state.lastLogIdx = len(n.state.logEntries) - 1
			break
		}
	}
	// pack resp
	resp.Success = true
	// append entries
	if len(req.entries) > 0 {
		n.state.logEntries = append(n.state.logEntries, req.entries[idx:]...)
		n.state.lastLogIdx = len(n.state.logEntries) - 1
	}
	// update node's commitIdx
	if req.leaderCommit > n.state.commitIdx {
		min := int(math.Min(float64(req.leaderCommit), float64(n.state.lastLogIdx)))
		n.state.commitIdx = min
	}
	return
}
