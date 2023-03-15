package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kennethfeh/system-monitor/internal/collector"
	"github.com/kennethfeh/system-monitor/internal/storage"
)

func TestNewServer(t *testing.T) {
	col := collector.NewCollector()
	stor := storage.NewMetricsStorage(10)
	
	server := NewServer(col, stor)
	
	if server == nil {
		t.Fatal("Expected server to be created")
	}
	
	if server.collector != col {
		t.Error("Expected collector to be set")
	}
	
	if server.storage != stor {
		t.Error("Expected storage to be set")
	}
	
	if server.clients == nil {
		t.Error("Expected clients map to be initialized")
	}
}

func TestHandleAPIMetrics(t *testing.T) {
	col := collector.NewCollector()
	stor := storage.NewMetricsStorage(10)
	server := NewServer(col, stor)
	
	req, err := http.NewRequest("GET", "/api/metrics", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleAPIMetrics)
	
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Handler returned wrong content type: got %v want %v",
			contentType, "application/json")
	}
}

func TestHandleAPIHistory(t *testing.T) {
	col := collector.NewCollector()
	stor := storage.NewMetricsStorage(10)
	server := NewServer(col, stor)
	
	// Add some test data
	metrics, _ := col.Collect()
	stor.Add(metrics)
	
	req, err := http.NewRequest("GET", "/api/history", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.handleAPIHistory)
	
	handler.ServeHTTP(rr, req)
	
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Handler returned wrong content type: got %v want %v",
			contentType, "application/json")
	}
}

func TestMetricsCollection(t *testing.T) {
	col := collector.NewCollector()
	stor := storage.NewMetricsStorage(10)
	server := NewServer(col, stor)
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	go server.startMetricsCollection(ctx)
	
	// Wait for at least one collection cycle
	time.Sleep(2500 * time.Millisecond)
	
	history := stor.GetHistory()
	if len(history) == 0 {
		t.Error("Expected at least one metric to be collected")
	}
}

func TestRoutes(t *testing.T) {
	router := mux.NewRouter()
	
	col := collector.NewCollector()
	stor := storage.NewMetricsStorage(10)
	server := NewServer(col, stor)
	
	router.HandleFunc("/api/metrics", server.handleAPIMetrics).Methods("GET")
	router.HandleFunc("/api/history", server.handleAPIHistory).Methods("GET")
	router.HandleFunc("/ws", server.handleWebSocket)
	
	tests := []struct {
		name     string
		method   string
		path     string
		expected int
	}{
		{"API Metrics", "GET", "/api/metrics", http.StatusOK},
		{"API History", "GET", "/api/history", http.StatusOK},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			
			if status := rr.Code; status != tt.expected {
				t.Errorf("Handler returned wrong status code: got %v want %v",
					status, tt.expected)
			}
		})
	}
}