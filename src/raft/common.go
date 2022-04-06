package raft

import (
	"math"
	"math/rand"
	"time"
)

const (
	MinElectionTimeout = 500
	MaxElectionTimeout = 1000
)

func randElectionTimeout() time.Duration {
	randTimeout := MinElectionTimeout + rand.Intn(MaxElectionTimeout-MinElectionTimeout)
	return time.Duration(randTimeout) * time.Millisecond
}

func minInt(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}
