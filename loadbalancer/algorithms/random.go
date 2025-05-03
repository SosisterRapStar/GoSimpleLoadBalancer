package algorithms

import (
	"math/rand"
	"sync"
	"time"

	"github.com/SosisterRapStar/GoSimpleLoadBalancer/core"
)

type Random struct {
	sync.Mutex
	upstream *core.Upstream
	random   *rand.Rand
}

func (r *Random) next() int {
	attempts := 0
	backendsNum := r.upstream.BackendsNum

	for attempts <= backendsNum {
		idx := r.random.Intn(backendsNum)
		if r.upstream.IsAvailable(idx) {
			return idx
		}
		attempts++
	}
	return -1
}

func (r *Random) Balance() int {
	r.Lock()
	defer r.Unlock()
	return r.next()
}

func NewRandom(upstream *core.Upstream) *Random {
	source := rand.NewSource(time.Now().UnixNano())
	return &Random{
		upstream: upstream,
		random:   rand.New(source),
	}
}
