package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Server struct {
	server *http.Server
}

func NewServer(addr string, loadbalancer http.Handler) *Server {

	httpServer := &http.Server{
		Addr:           addr,
		Handler:        loadbalancer,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: (1 << 20),
	}
	server := &Server{
		server: httpServer,
	}
	return server
}

func (s *Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		logger.Info(fmt.Sprintf("Started to listen on %s", s.server.Addr))
	}()

}

func (s *Server) Stop(ctx context.Context) {
	logger.Info(fmt.Sprintf("Shutting down the server on %s", s.server.Addr))
	if err := s.server.Shutdown(ctx); err != nil {
		log.Fatal("Error occured shutting the server down")
	}
	logger.Info("Balancer was stoped")
}
