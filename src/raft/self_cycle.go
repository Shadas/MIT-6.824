package raft

import (
	"time"
)

func (n *Node) run() {
	for {
		n.Lock()
		role := n.state.role
		n.Unlock()
		switch role {
		case FollowerRole:
			n.runAsFollower()
		case CandidateRole:
			n.runAsCandidate()
		case LeaderRole:
			n.runAsLeader()
		default:
			panic("invalid node role")
		}
		time.Sleep(time.Second)
	}
}

func (n *Node) runAsFollower() {
	dur := randElectionTimeout()
	time.Sleep(dur)
	n.Lock()
	lastAccessed := n.state.lastAccessed
	n.Unlock()
	if time.Now().Sub(lastAccessed).Milliseconds() >= dur.Milliseconds() {
		n.Lock()
		n.state.role = CandidateRole
		n.state.currentTerm++
		n.state.voteFor = -1
		n.Unlock()
	}

}

func (n *Node) sendRequestVote(server int, req *RequestVoteReq) (resp *RequestVoteResp, ok bool) {
	node := n.peers[server]
	resp = node.RequestVote(req)
	ok = true // todo: add ok logic
	return
}

func (n *Node) runAsCandidate() {
	dur := randElectionTimeout()
	start := time.Now()
	n.Lock()
	peers := n.peers
	me := n.me
	term := n.state.currentTerm
	lastLogIdx := n.state.lastLogIdx
	lastLogTerm := n.state.logEntries[lastLogIdx].Term
	n.Unlock()
	var (
		count    = 0
		total    = len(peers)
		finished = 0
		majority = total/2 + 1
	)
	for peer := range peers {
		if me == peer { // range with the node self
			n.Lock()
			count++
			finished++
			n.Unlock()
			continue
		}
		go func(peer int) {
			req := &RequestVoteReq{
				CandidateTerm: term,
				CandidateId:   me,
				LastLogIdx:    lastLogIdx,
				LastLogTerm:   lastLogTerm,
			}
			resp, ok := n.sendRequestVote(peer, req)
			n.Lock()
			defer n.Unlock()
			if !ok {
				finished++
				return
			}
			if resp.VoteGranted {
				finished++
				count++
			} else {
				finished++
				if req.CandidateTerm < resp.Term {
					n.state.role = FollowerRole
				}
			}
		}(peer)
	}
	for {
		n.Lock()
		if count >= majority || finished == total || time.Now().Sub(start).Milliseconds() >= dur.Milliseconds() {
			break
		}
		n.Unlock()
		time.Sleep(time.Millisecond * 50)
	}
	if time.Now().Sub(start).Milliseconds() >= dur.Milliseconds() {
		n.state.role = FollowerRole
		n.Unlock()
		return
	}
	if n.state.role == CandidateRole && count >= majority {
		n.state.role = LeaderRole
		for peer := range peers {
			n.state.nextIndex[peer] = n.state.lastLogIdx + 1
		}
	} else {
		n.state.role = FollowerRole
	}
	n.Unlock()
}

func (n *Node) sendAppendEntries(server int, req *AppendEntriesReq) (resp *AppendEntriesResp, ok bool) {
	node := n.peers[server]
	resp = node.AppendEntries(req)
	ok = true // todo: add ok logic
	return
}

func (n *Node) runAsLeader() {
	n.Lock()
	me := n.me
	term := n.state.currentTerm
	commitIdx := n.state.commitIdx
	peers := n.peers
	nextIdx := n.state.nextIndex
	lastLogIdx := n.state.lastLogIdx
	matchIdx := n.state.matchIndex
	nextIdx[me] = lastLogIdx + 1
	matchIdx[me] = lastLogIdx
	logs := n.state.logEntries
	n.Unlock()

	for i := commitIdx + 1; i <= lastLogIdx; i++ {
		count := 0
		total := len(peers)
		majority := (total / 2) + 1
		for peer := range peers {
			if matchIdx[peer] >= i && logs[i].Term == term {
				count++
			}
		}

		if count >= majority {
			n.Lock()
			for j := commitIdx + 1; j <= i; j++ {
				n.state.commitIdx = n.state.commitIdx + 1
			}
			n.Unlock()
		}
	}

	for peer := range peers {
		if peer == me {
			continue
		}
		req := &AppendEntriesReq{}
		n.Lock()
		req.Term = n.state.currentTerm
		req.NodeId = me
		prevLogIdx := nextIdx[peer] - 1
		req.prevLogIdx = prevLogIdx
		req.prevLogTerm = n.state.logEntries[prevLogIdx].Term
		req.leaderCommit = n.state.commitIdx
		if nextIdx[peer] <= lastLogIdx {
			req.entries = n.state.logEntries[prevLogIdx+1 : lastLogIdx+1]
		}
		n.Unlock()

		go func(peer int) {
			resp, ok := n.sendAppendEntries(peer, req)
			if !ok {
				return
			}

			n.Lock()
			if resp.Success {
				n.state.nextIndex[peer] = minInt(n.state.nextIndex[peer]+len(req.entries), n.state.lastLogIdx+1)
				n.state.matchIndex[peer] = prevLogIdx + len(req.entries)
			} else {
				if resp.Term > req.Term {
					n.state.role = FollowerRole
					n.Unlock()
					return
				}
			}
			n.Unlock()
		}(peer)
	}
}
