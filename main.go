package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kennethfeh/system-monitor/internal/collector"
	"github.com/kennethfeh/system-monitor/internal/models"
	"github.com/kennethfeh/system-monitor/internal/storage"
)

//go:embed static/*
var staticFiles embed.FS

//go:embed templates/*
var templateFiles embed.FS

var (
	port     = flag.String("port", "8080", "Port to run the server on")
	interval = flag.Duration("interval", 2*time.Second, "Metrics collection interval")
	history  = flag.Int("history", 60, "Number of historical data points to keep")
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	collector *collector.Collector
	storage   *storage.MetricsStorage
	clients   map[*websocket.Conn]bool
	broadcast chan models.SystemMetrics
	register  chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewServer(collector *collector.Collector, storage *storage.MetricsStorage) *Server {
	return &Server{
		collector:  collector,
		storage:    storage,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan models.SystemMetrics),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

func (s *Server) run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client] = true
			log.Println("Client connected")
			
			// Send historical data to new client
			history := s.storage.GetHistory()
			for _, metrics := range history {
				if err := client.WriteJSON(metrics); err != nil {
					log.Printf("Error sending historical data: %v", err)
					break
				}
			}

		case client := <-s.unregister:
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				client.Close()
				log.Println("Client disconnected")
			}

		case metrics := <-s.broadcast:
			for client := range s.clients {
				err := client.WriteJSON(metrics)
				if err != nil {
					log.Printf("Error writing to client: %v", err)
					client.Close()
					delete(s.clients, client)
				}
			}
		}
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	s.register <- conn
	
	defer func() {
		s.unregister <- conn
	}()
	
	// Keep connection alive and handle ping/pong
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(templateFiles, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	data := struct {
		Title string
		Port  string
	}{
		Title: "System Monitor",
		Port:  *port,
	}
	
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := s.collector.Collect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (s *Server) handleAPIHistory(w http.ResponseWriter, r *http.Request) {
	history := s.storage.GetHistory()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (s *Server) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(*interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := s.collector.Collect()
			if err != nil {
				log.Printf("Error collecting metrics: %v", err)
				continue
			}
			
			s.storage.Add(metrics)
			s.broadcast <- metrics
		}
	}
}

func main() {
	flag.Parse()
	
	log.Printf("Starting System Monitor on port %s", *port)
	log.Printf("Collection interval: %v", *interval)
	log.Printf("History size: %d data points", *history)
	
	// Initialize components
	collector := collector.NewCollector()
	storage := storage.NewMetricsStorage(*history)
	server := NewServer(collector, storage)
	
	// Start WebSocket handler
	go server.run()
	
	// Start metrics collection
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go server.startMetricsCollection(ctx)
	
	// Setup routes
	router := mux.NewRouter()
	
	// API routes
	router.HandleFunc("/api/metrics", server.handleAPIMetrics).Methods("GET")
	router.HandleFunc("/api/history", server.handleAPIHistory).Methods("GET")
	router.HandleFunc("/ws", server.handleWebSocket)
	
	// Static files
	router.PathPrefix("/static/").Handler(http.FileServer(http.FS(staticFiles)))
	
	// Main page
	router.HandleFunc("/", server.handleIndex).Methods("GET")
	
	// Setup HTTP server
	srv := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		
		log.Println("Shutting down server...")
		cancel()
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()
	
	// Start server
	fmt.Printf("\nSystem Monitor is running at http://localhost:%s\n", *port)
	fmt.Println("Press Ctrl+C to stop")
	
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
	
	log.Println("Server stopped")
}