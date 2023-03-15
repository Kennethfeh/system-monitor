package storage

import (
	"testing"
	"time"

	"github.com/kennethfeh/system-monitor/internal/models"
)

func TestNewMetricsStorage(t *testing.T) {
	tests := []struct {
		name     string
		maxSize  int
		expected int
	}{
		{"Positive size", 10, 10},
		{"Zero size", 0, 60},
		{"Negative size", -5, 60},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMetricsStorage(tt.maxSize)
			
			if storage == nil {
				t.Fatal("Expected storage to be created")
			}
			
			if storage.maxSize != tt.expected {
				t.Errorf("Expected maxSize to be %d, got %d", tt.expected, storage.maxSize)
			}
			
			if storage.metrics == nil {
				t.Error("Expected metrics slice to be initialized")
			}
		})
	}
}

func TestAdd(t *testing.T) {
	storage := NewMetricsStorage(3)
	
	// Add first metric
	metric1 := models.SystemMetrics{
		Timestamp: time.Now(),
	}
	storage.Add(metric1)
	
	if storage.Size() != 1 {
		t.Errorf("Expected size to be 1, got %d", storage.Size())
	}
	
	// Add second metric
	metric2 := models.SystemMetrics{
		Timestamp: time.Now().Add(1 * time.Second),
	}
	storage.Add(metric2)
	
	if storage.Size() != 2 {
		t.Errorf("Expected size to be 2, got %d", storage.Size())
	}
	
	// Add third metric
	metric3 := models.SystemMetrics{
		Timestamp: time.Now().Add(2 * time.Second),
	}
	storage.Add(metric3)
	
	if storage.Size() != 3 {
		t.Errorf("Expected size to be 3, got %d", storage.Size())
	}
	
	// Add fourth metric (should remove oldest)
	metric4 := models.SystemMetrics{
		Timestamp: time.Now().Add(3 * time.Second),
	}
	storage.Add(metric4)
	
	if storage.Size() != 3 {
		t.Errorf("Expected size to remain 3, got %d", storage.Size())
	}
	
	// Verify oldest metric was removed
	history := storage.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected history length to be 3, got %d", len(history))
	}
	
	// First metric should be metric2 now
	if !history[0].Timestamp.Equal(metric2.Timestamp) {
		t.Error("Expected oldest metric to be removed")
	}
}

func TestGetHistory(t *testing.T) {
	storage := NewMetricsStorage(5)
	
	// Test empty history
	history := storage.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(history))
	}
	
	// Add some metrics
	for i := 0; i < 3; i++ {
		metric := models.SystemMetrics{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		storage.Add(metric)
	}
	
	history = storage.GetHistory()
	if len(history) != 3 {
		t.Errorf("Expected 3 items in history, got %d", len(history))
	}
	
	// Verify history is a copy (modifying it shouldn't affect storage)
	if len(history) > 0 {
		originalTime := history[0].Timestamp
		history[0].Timestamp = time.Now().Add(1 * time.Hour)
		
		newHistory := storage.GetHistory()
		if !newHistory[0].Timestamp.Equal(originalTime) {
			t.Error("History should be a copy, not a reference")
		}
	}
}

func TestGetLatest(t *testing.T) {
	storage := NewMetricsStorage(5)
	
	// Test with empty storage
	latest := storage.GetLatest()
	if latest != nil {
		t.Error("Expected nil for empty storage")
	}
	
	// Add metrics
	metric1 := models.SystemMetrics{
		Timestamp: time.Now(),
		System: models.SystemInfo{
			Hostname: "host1",
		},
	}
	storage.Add(metric1)
	
	metric2 := models.SystemMetrics{
		Timestamp: time.Now().Add(1 * time.Second),
		System: models.SystemInfo{
			Hostname: "host2",
		},
	}
	storage.Add(metric2)
	
	latest = storage.GetLatest()
	if latest == nil {
		t.Fatal("Expected latest metric to be returned")
	}
	
	if latest.System.Hostname != "host2" {
		t.Errorf("Expected latest metric to have hostname 'host2', got '%s'", latest.System.Hostname)
	}
	
	// Verify it's a copy
	latest.System.Hostname = "modified"
	newLatest := storage.GetLatest()
	if newLatest.System.Hostname != "host2" {
		t.Error("Latest should be a copy, not a reference")
	}
}

func TestClear(t *testing.T) {
	storage := NewMetricsStorage(5)
	
	// Add some metrics
	for i := 0; i < 3; i++ {
		metric := models.SystemMetrics{
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		storage.Add(metric)
	}
	
	if storage.Size() != 3 {
		t.Errorf("Expected size to be 3, got %d", storage.Size())
	}
	
	// Clear storage
	storage.Clear()
	
	if storage.Size() != 0 {
		t.Errorf("Expected size to be 0 after clear, got %d", storage.Size())
	}
	
	history := storage.GetHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d items", len(history))
	}
	
	latest := storage.GetLatest()
	if latest != nil {
		t.Error("Expected nil latest after clear")
	}
}

func TestConcurrentAccess(t *testing.T) {
	storage := NewMetricsStorage(100)
	done := make(chan bool)
	
	// Writer goroutine
	go func() {
		for i := 0; i < 50; i++ {
			metric := models.SystemMetrics{
				Timestamp: time.Now(),
			}
			storage.Add(metric)
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()
	
	// Reader goroutine 1
	go func() {
		for i := 0; i < 50; i++ {
			_ = storage.GetHistory()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()
	
	// Reader goroutine 2
	go func() {
		for i := 0; i < 50; i++ {
			_ = storage.GetLatest()
			_ = storage.Size()
			time.Sleep(1 * time.Millisecond)
		}
		done <- true
	}()
	
	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
	
	// Verify storage is still consistent
	size := storage.Size()
	history := storage.GetHistory()
	if len(history) != size {
		t.Errorf("Inconsistent state: size=%d, history length=%d", size, len(history))
	}
}