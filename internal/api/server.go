package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/orchard9/watch-now/internal/core"
	"github.com/orchard9/watch-now/internal/monitors"
)

type Server struct {
	engine   *core.Engine
	server   *http.Server
	listener net.Listener
}

type StatusResponse struct {
	Timestamp string                      `json:"timestamp"`
	Services  []*monitors.Result          `json:"services"`
	Checks    []*monitors.Result          `json:"checks"`
	Overall   string                      `json:"overall"`
	Results   map[string]*monitors.Result `json:"results"`
}

func NewServer(engine *core.Engine, port int) *Server {
	s := &Server{
		engine: engine,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/events", s.handleSSE)
	mux.HandleFunc("/api/health", s.handleHealth)

	s.server = &http.Server{
		Handler:      s.corsMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Create listener
	var err error
	addr := fmt.Sprintf(":%d", port)
	s.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("API server starting on http://localhost:%d", s.listener.Addr().(*net.TCPAddr).Port)
	return s.server.Serve(s.listener)
}

func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

func (s *Server) Port() int {
	if s.listener != nil {
		return s.listener.Addr().(*net.TCPAddr).Port
	}
	return 0
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	results := s.engine.State().GetAll()

	// Group results by type
	var services []*monitors.Result
	var checks []*monitors.Result

	for _, result := range results {
		switch result.Type {
		case monitors.TypeQuality:
			checks = append(checks, result)
		case monitors.TypeREST, monitors.TypeGRPC:
			services = append(services, result)
		}
	}

	// Determine overall status
	overall := s.getOverallStatus(results)

	response := StatusResponse{
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Services:  services,
		Checks:    checks,
		Overall:   string(overall),
		Results:   results,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a channel to receive state updates
	updates := make(chan map[string]*monitors.Result, 10)
	s.engine.State().Subscribe(updates)
	defer s.engine.State().Unsubscribe(updates)

	// Send initial state
	s.sendSSEEvent(w, "status", s.getStatusData())

	// Set up ticker for periodic updates
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Handle client disconnect
	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-updates:
			// Send updated status when state changes
			s.sendSSEEvent(w, "status", s.getStatusData())
		case <-ticker.C:
			// Send periodic heartbeat
			s.sendSSEEvent(w, "heartbeat", map[string]interface{}{
				"timestamp": time.Now().Unix(),
			})
		}

		// Flush the response
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func (s *Server) sendSSEEvent(w http.ResponseWriter, event string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling SSE data: %v", err)
		return
	}

	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonData))
}

func (s *Server) getStatusData() StatusResponse {
	results := s.engine.State().GetAll()

	// Group results by type
	var services []*monitors.Result
	var checks []*monitors.Result

	for _, result := range results {
		switch result.Type {
		case monitors.TypeQuality:
			checks = append(checks, result)
		case monitors.TypeREST, monitors.TypeGRPC:
			services = append(services, result)
		}
	}

	// Determine overall status
	overall := s.getOverallStatus(results)

	return StatusResponse{
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Services:  services,
		Checks:    checks,
		Overall:   string(overall),
		Results:   results,
	}
}

func (s *Server) getOverallStatus(results map[string]*monitors.Result) monitors.Status {
	if len(results) == 0 {
		return monitors.StatusInfo
	}

	hasWarn := false
	for _, result := range results {
		if result.Status == monitors.StatusFail {
			return monitors.StatusFail
		}
		if result.Status == monitors.StatusWarn {
			hasWarn = true
		}
	}

	if hasWarn {
		return monitors.StatusWarn
	}
	return monitors.StatusOK
}
