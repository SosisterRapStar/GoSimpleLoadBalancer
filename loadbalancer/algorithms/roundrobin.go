package algorithms

import (
	"sync"

	"github.com/SosisterRapStar/GoSimpleLoadBalancer/core"
)

type RoundRobin struct {
	sync.Mutex
	upstream          *core.Upstream
	roundRobinClosure func() int
}

func (rr *RoundRobin) next() int {
	attempts := 0
	var next int
	for attempts <= rr.upstream.BackendsNum {
		next = rr.roundRobinClosure()
		if rr.upstream.IsAvailable(next) {
			return next
		} else {
			attempts += 1
		}
	}
	return -1
}

func (rr *RoundRobin) Balance() int {
	rr.Lock()
	defer rr.Unlock()
	return rr.next()
}

func NewRoundRobin(upstream *core.Upstream) *RoundRobin {
	return &RoundRobin{
		upstream:          upstream,
		roundRobinClosure: roundRobin(len(upstream.GetInfoAllBackends())),
	}
}

func roundRobin(servers_num int) func() int {
	var (
		x int = -1
	)
	return func() int {
		x += 1
		x %= servers_num
		return x
	}
}
