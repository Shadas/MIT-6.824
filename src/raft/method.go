package raft

type RequestVoteReq struct {
	CandidateTerm int
	CandidateId   int
	LastLogIdx    int
	LastLogTerm   int
}

type RequestVoteResp struct {
	Term        int
	VoteGranted bool
}

func (n *Node) RequestVote(req *RequestVoteReq) (resp *RequestVoteResp) {
	resp = &RequestVoteResp{
		VoteGranted: false,
		Term:        n.state.currentTerm,
	}
	// return true if the request from the same node
	if req.CandidateTerm == n.state.currentTerm && req.CandidateId == n.state.voteFor {
		resp.VoteGranted = true
		resp.Term = n.state.currentTerm
		return
	}
	// ignore the old term
	if req.CandidateTerm < n.state.currentTerm {
		return
	}
	// the node has voted for other node in the same term.
	if n.state.currentTerm == req.CandidateTerm && n.state.voteFor != 0 && n.state.voteFor != req.CandidateId {
		return
	}
	// a bigger term
	if req.CandidateTerm > n.state.currentTerm {
		n.state.currentTerm = req.CandidateTerm
		n.state.role = FollowerRole
	}
	// use new candidate term
	resp.Term = req.CandidateTerm
	// check log version
	if n.state.lastLogIdx-1 >= 0 {
		lastLogTerm := n.state.logEntries[n.state.lastLogIdx-1].Term
		if lastLogTerm > req.LastLogTerm || // request with old term
			(lastLogTerm == req.LastLogTerm && n.state.lastLogIdx > req.LastLogIdx) { // request with the same term and old idx
			resp.VoteGranted = false
			return
		}
	}
	// set local node
	n.state.role = FollowerRole
	resp.VoteGranted = true
	n.state.voteFor = req.CandidateId
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
		min := minInt(req.leaderCommit, n.state.lastLogIdx)
		n.state.commitIdx = min
	}
	return
}
