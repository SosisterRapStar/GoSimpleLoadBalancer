package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SosisterRapStar/GoSimpleLoadBalancer/algorithms"
	"github.com/SosisterRapStar/GoSimpleLoadBalancer/core"
	"github.com/SosisterRapStar/GoSimpleLoadBalancer/healthcheck"
	"github.com/SosisterRapStar/GoSimpleLoadBalancer/logging"
	"github.com/SosisterRapStar/GoSimpleLoadBalancer/readers"
)

var logger = logging.GetLogger()

func main() {
	c := &readers.Config{}
	err := readers.ReadConfig(c)
	if err != nil {
		log.Fatal(err)
	}
	if c.Listen == "" {
		log.Fatal("You need to specify listen address")
	}
	if len(c.Upstreams) == 0 {
		log.Fatal("You need to specify upstreams")
	}

	ctx, cancel := context.WithCancel(context.Background())

	var backends []*core.Backend
	for _, addr := range c.Upstreams {
		backends = append(backends, core.NewBackend(addr, ""))
	}

	upstream := core.NewUpstream(backends, "Configured upstream")

	var healthchecker *healthcheck.DefaultHealthChecker
	if c.Healthcheck == nil {
		log.Fatal("Without healthcheck it will be hard to detect alive backends, so you need to specify it")
	}
	if c.Healthcheck != nil {
		if c.Healthcheck.Endpoint == nil {
			log.Fatal("Healthcheck endpoint should be specified")
		}
		healthchecker = healthcheck.NewDefaultHealthChecker(
			getIntWithDefault(c.Healthcheck.Period, 10),
			getIntWithDefault(c.Healthcheck.ResponseTimeoutSeconds, 10),
			getIntWithDefault(c.Healthcheck.MaxRetries, 3),
			getIntWithDefault(c.Healthcheck.TimeOutStep, 2),
			*c.Healthcheck.Endpoint,
			upstream)
	}

	updates := make(chan *core.HealthStatus)

	go healthchecker.StartCheck(ctx, updates)
	go upstream.BackendsMonitoring(ctx, updates)

	var lbAlgorithm algorithms.BalanceAlgorithm
	switch c.Balance {
	case "roundrobin":
		lbAlgorithm = algorithms.NewRoundRobin(upstream)
	case "random":
		lbAlgorithm = algorithms.NewRandom(upstream)
	default:
		logger.Info("Can not recognize balance algo use roundrobin by default")
		lbAlgorithm = algorithms.NewRoundRobin(upstream)
	}

	responseHeaderTimeout := 10
	if c.Balancer != nil && c.Balancer.ResponseHeaderTimeout != nil {
		responseHeaderTimeout = *c.Balancer.ResponseHeaderTimeout
	}

	loadBalancer := NewLoadBalancer(upstream, lbAlgorithm, responseHeaderTimeout)

	fmt.Println("Formed loadbalancer for upstreams:")
	for _, b := range upstream.GetInfoAllBackends() {
		fmt.Println(b.Addr)
	}

	server := NewServer(c.Listen, loadBalancer)
	server.Start()

	fmt.Println("Simple loadbalancer started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Debug("Closing main updates channel")
	close(updates)
	cancel()
	logger.Info("Received Interrupt signal, started to shutdown gracefully")
	gshutCtx, gshutClose := context.WithTimeout(context.Background(), 10*time.Second)
	defer gshutClose()
	server.Stop(gshutCtx)
}

func getIntWithDefault(val *int, defaultVal int) int {
	if val != nil {
		return *val
	}
	return defaultVal
}
