package raft

type Node struct {
	state *State
}

func (n *Node) RequestVote(candidateTerm, candidateId, lastLogIdx, lastLogTerm int64) (term int64, voteGranted bool) {
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
	Term         int64 // leader's term
	NodeId       int64 // leader's id
	prevLogIdx   int64
	prevLogTerm  int64
	leaderCommit int64
	entries      []*LogEntry
}

type AppendEntriesResp struct {
	Term    int64
	Success bool
}

func (n *Node) AppendEntries(req *AppendEntriesReq) (resp *AppendEntriesResp) {
	resp = &AppendEntriesResp{
		Term:    -1,
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
	if len(n.state.logEntries) < int(req.prevLogIdx+1) {
		return
	}
	// check prev log idx if needed
	if len(n.state.logEntries) > int(req.prevLogIdx) {
		l := n.state.logEntries[prevLogIdx]
		if l.Term != prevLogTerm {
			success = false
			return
		}
	}
	// remove conflict logs
	for _, e := range entries {

	}
	return
}
