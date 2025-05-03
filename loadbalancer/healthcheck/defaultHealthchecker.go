package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/SosisterRapStar/GoSimpleLoadBalancer/core"
	"github.com/SosisterRapStar/GoSimpleLoadBalancer/logging"
)

var logger = logging.GetLogger()

// так как логика отдельных алгоритмов healthcheck'а может быть разной, каждый алгоритм будет иметь свою структуру и хранить в ней свою метадату
type DefaultHealthData struct {
	sync.Mutex
	RetriesCounter int
	CurrentDelay   time.Duration
	LastFailure    time.Time
}

type DefaultHealthChecker struct {
	sync.Mutex
	upstreamName             string
	backendInfo              []core.BackendInfo
	healthParams             []*DefaultHealthData
	connectionTimeoutSeconds int
	checkEndpoint            string
	timeoutStep              int
	maxRetries               int
	checkIntervalSeconds     int
	httpClient               http.Client
}

func NewDefaultHealthChecker(checkIntervalSeconds int, timeoutSeconds int, maxRetries int, timeoutStep int, checkEndpoint string, upstream *core.Upstream) *DefaultHealthChecker {
	backendInfo := upstream.GetInfoAllBackends()
	healthParams := make([]*DefaultHealthData, len(backendInfo))
	for i := range healthParams {
		healthParams[i] = &DefaultHealthData{}
	}
	return &DefaultHealthChecker{
		upstreamName:             upstream.Name,
		healthParams:             healthParams,
		backendInfo:              backendInfo,
		connectionTimeoutSeconds: timeoutSeconds,
		checkEndpoint:            checkEndpoint,
		maxRetries:               maxRetries,
		timeoutStep:              timeoutStep,
		checkIntervalSeconds:     checkIntervalSeconds,
		httpClient:               http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second},
	}
}

func (h *DefaultHealthChecker) isReadyForCheck(index int) bool {
	h.healthParams[index].Lock()
	defer h.healthParams[index].Unlock()
	return time.Now().After(h.healthParams[index].LastFailure.Add(h.healthParams[index].CurrentDelay))
}

func (h *DefaultHealthChecker) check(index int, url string, out chan<- *core.HealthStatus) {

	// logger.Debug(fmt.Sprintf("Start checkign %s", url))
	if !h.isReadyForCheck(index) {
		out <- &core.HealthStatus{Index: index, IsAvailable: false}
		return
	}
	resp, err := h.httpClient.Get(url)
	// logger.Debug(fmt.Sprintf("Response for %s %s", url, resp.Status))
	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		out <- &core.HealthStatus{Index: index, IsAvailable: false}
		return
	}
	out <- &core.HealthStatus{Index: index, IsAvailable: true}
}

func (h *DefaultHealthChecker) urlConstructor(addr string) string {

	return fmt.Sprintf("http://%s%s", addr, h.checkEndpoint)
}

func (h *DefaultHealthChecker) timeoutAlgo(index int, isAvailable bool) {
	h.healthParams[index].Lock()
	defer h.healthParams[index].Unlock()
	if !isAvailable {
		h.healthParams[index].LastFailure = time.Now()
		if h.healthParams[index].RetriesCounter < h.maxRetries {
			h.healthParams[index].RetriesCounter += 1
			h.healthParams[index].CurrentDelay = time.Duration(h.timeoutStep * h.healthParams[index].RetriesCounter)
		} else {
			h.healthParams[index].CurrentDelay = time.Duration(h.timeoutStep * h.healthParams[index].RetriesCounter)
		}

	} else {
		h.healthParams[index].LastFailure = time.Time{}
		h.healthParams[index].RetriesCounter = 0
		h.healthParams[index].CurrentDelay = time.Duration(h.timeoutStep * h.healthParams[index].RetriesCounter)
	}
}

func (h *DefaultHealthChecker) StartCheck(ctx context.Context, updates chan<- *core.HealthStatus) {
	outChanForCheck := make(chan *core.HealthStatus)
	ticker := time.NewTicker(time.Duration(h.checkIntervalSeconds) * time.Second)
	defer ticker.Stop()
	defer h.httpClient.CloseIdleConnections()
	logchan := make(chan HealthLogStruct, len(h.backendInfo))

	go StartHealthLogger(logchan)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Debug("Close healthcheck demultiplexor")
				return
			case status := <-outChanForCheck:
				h.timeoutAlgo(status.Index, status.IsAvailable)
				updates <- status
				logchan <- HealthLogStruct{
					Addr:    h.backendInfo[status.Index].Addr,
					Url:     h.backendInfo[status.Index].ProxyUrl.String(),
					IsAlive: status.IsAvailable,
				}
			}
		}
	}()

	var wg sync.WaitGroup

loop:
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Close healthcheck multiplexor")
			break loop
		case <-ticker.C:
			for i, backend := range h.backendInfo {
				wg.Add(1)
				go func(i int, addr string) {
					defer wg.Done()
					h.check(i, addr, outChanForCheck)
				}(i, h.urlConstructor(backend.Addr))
			}
		}
	}

	wg.Wait()
	close(logchan)
	close(outChanForCheck)

}
