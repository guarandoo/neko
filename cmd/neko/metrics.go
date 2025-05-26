package main

import (
	"log"
	"net/http"
	"sync"
)

type MetricsServer interface {
	Listen(addr string) error
	Close() error
}

type metricsServer struct {
	mux     *http.ServeMux
	server  *http.Server
	wg      sync.WaitGroup
	running bool
}

func (s *metricsServer) Listen(addr string) error {
	if s.running {
		return nil
	}

	log.Printf("starting metrics server on %v", addr)

	server := http.Server{
		Addr:    addr,
		Handler: s.mux,
	}
	s.server = &server

	s.running = true
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer func() {
			s.running = false
			s.server = nil
		}()

		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("unable to start metrics server: %v", err)
		}

		log.Printf("metrics server terminated")
	}()

	return nil
}

func (s *metricsServer) Close() error {
	if !s.running {
		return nil
	}

	if err := s.server.Close(); err != nil {
		return err
	}

	s.wg.Wait()

	return nil
}

func newMetricsServer(handler http.Handler) MetricsServer {
	mux := http.NewServeMux()
	mux.Handle("/metrics", handler)

	server := metricsServer{
		mux: mux,
	}

	return &server
}
