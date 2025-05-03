package core

import (
	"context"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
)

// Домен бэкенда и апстрима

type BackendInfo struct {
	Addr string `json:"Address"`
	// NumOfCons   int      `json:"CurrentActiveCons"`
	ProxyUrl *url.URL `json:"ProxyUrl"`
}

type Backend struct {
	addr        string
	isAvailable int32
	// numOfCons   int
	proxyUrl *url.URL
}

func (b *Backend) SetAvailability(status bool) {
	if status {
		atomic.StoreInt32(&b.isAvailable, 1)
		return
	}
	atomic.StoreInt32(&b.isAvailable, 0)

}

func (b *Backend) IsAvailable() bool {
	return atomic.LoadInt32(&b.isAvailable) == 1
}

// Upstream - это агрегат для бэкендов, все получения бэкендов только через него, иначе не получится контролировать конкурентный доступ к бэкендам
type Upstream struct {
	Name         string
	BackendsNum  int
	backends     []*Backend
	backendlocks []*sync.RWMutex
}

func (u *Upstream) IsAvailable(index int) bool {
	return u.backends[index].IsAvailable()
}

func (u *Upstream) GetBackendInfo(index int) BackendInfo {

	u.backendlocks[index].RLock() // Используем RLock для чтения
	defer u.backendlocks[index].RUnlock()
	b := u.backends[index]

	return BackendInfo{
		Addr: b.addr,
		// NumOfCons:   b.numOfCons,
		ProxyUrl: b.proxyUrl,
	}
}

func (u *Upstream) GetInfoAllBackends() []BackendInfo {
	resp := make([]BackendInfo, len(u.backends))
	for i := range resp {
		resp[i] = u.GetBackendInfo(i)
	}
	return resp
}

func (u *Upstream) BackendsMonitoring(ctx context.Context, updates <-chan *HealthStatus) {
	for {
		select {
		case newStatus, ok := <-updates:
			if !ok {
				return
			}
			b := u.backends[newStatus.Index]
			b.SetAvailability(newStatus.IsAvailable)
		case <-ctx.Done():
			return
		}
	}
}

func NewBackend(addr string, proxyPrefix string) *Backend {
	u, err := url.Parse("http://" + addr + proxyPrefix)
	if err != nil {
		log.Fatalf("Can not parse addr %s", "http://"+addr+proxyPrefix)
	}

	return &Backend{
		addr:        addr,
		isAvailable: 1,
		// numOfCons:   0,
		proxyUrl: u,
	}
}

func NewUpstream(backends []*Backend, name string) *Upstream {
	backendlocks := make([]*sync.RWMutex, len(backends))
	for i := range backendlocks {
		backendlocks[i] = &sync.RWMutex{}
	}
	return &Upstream{
		Name:         name,
		BackendsNum:  len(backends),
		backends:     backends,
		backendlocks: backendlocks,
	}
}
