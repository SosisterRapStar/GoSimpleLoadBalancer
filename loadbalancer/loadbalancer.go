package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/SosisterRapStar/GoSimpleLoadBalancer/algorithms"
	"github.com/SosisterRapStar/GoSimpleLoadBalancer/core"
)

type contextKey string

const proxyURLKey contextKey = "proxyURL"

type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (lrw *LoggingResponseWriter) WriteHeader(statusCode int) {
	lrw.StatusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

type LoadBalancer struct {
	sync.Mutex
	algorithm algorithms.BalanceAlgorithm

	upstream                 *core.Upstream
	RP                       *httputil.ReverseProxy
	connectionTimeOutSeconds int
}

func composeBestTransport(numOfBackends int, responseTimeOut int) *http.Transport {
	// параметры подобраны рандомно согласно некому вайбу

	return &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       numOfBackends * 2,
		ResponseHeaderTimeout: time.Duration(responseTimeOut) * time.Second,
		IdleConnTimeout:       90 * time.Second,
		ForceAttemptHTTP2:     true,
	}
}

func NewLoadBalancer(upstream *core.Upstream, balanceAlgo algorithms.BalanceAlgorithm, timeOut int) *LoadBalancer {

	lb := &LoadBalancer{
		upstream:                 upstream,
		algorithm:                balanceAlgo,
		connectionTimeOutSeconds: timeOut,
	}

	reverseProxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			proxyUrl := r.In.Context().Value(proxyURLKey).(*url.URL)
			r.SetURL(proxyUrl)
			r.Out.Host = r.In.Host
		},
		Transport: composeBestTransport(upstream.BackendsNum, lb.connectionTimeOutSeconds),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logger.Info(fmt.Sprintf("Proxy timeout during request to %s", r.Context().Value(proxyURLKey).(*url.URL).String()))
			}
		},
	}

	lb.RP = reverseProxy
	return lb

}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backendIndex := lb.algorithm.Balance()
	if backendIndex == -1 {
		http.Error(w, "No awailable servers, all backends are down", http.StatusInternalServerError)
		logger.Debug("No awailable servers, all backends are down")
		return
	}

	backendInfo := lb.upstream.GetBackendInfo(backendIndex)
	logger.Debug(fmt.Sprintf("Proxy request http://%s%s to %s", r.Host, r.URL.String(), backendInfo.ProxyUrl.String()))

	logwriter := &LoggingResponseWriter{
		ResponseWriter: w,
	}
	ctx := context.WithValue(context.Background(), proxyURLKey, backendInfo.ProxyUrl)
	rctx := r.WithContext(ctx)
	lb.RP.ServeHTTP(logwriter, rctx)
	logger.Debug(fmt.Sprintf("Returned request to %s with status: %d", backendInfo.ProxyUrl.String(), logwriter.StatusCode))
}
