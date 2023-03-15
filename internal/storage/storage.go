package storage

import (
	"sync"

	"github.com/kennethfeh/system-monitor/internal/models"
)

// MetricsStorage stores historical metrics data
type MetricsStorage struct {
	mu       sync.RWMutex
	metrics  []models.SystemMetrics
	maxSize  int
}

// NewMetricsStorage creates a new metrics storage with specified history size
func NewMetricsStorage(maxSize int) *MetricsStorage {
	if maxSize <= 0 {
		maxSize = 60 // Default to 60 data points
	}
	return &MetricsStorage{
		metrics: make([]models.SystemMetrics, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a new metric to the storage
func (s *MetricsStorage) Add(metric models.SystemMetrics) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = append(s.metrics, metric)
	
	// Remove oldest entries if we exceed max size
	if len(s.metrics) > s.maxSize {
		s.metrics = s.metrics[len(s.metrics)-s.maxSize:]
	}
}

// GetHistory returns all stored metrics
func (s *MetricsStorage) GetHistory() []models.SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modifications
	history := make([]models.SystemMetrics, len(s.metrics))
	copy(history, s.metrics)
	return history
}

// GetLatest returns the most recent metric
func (s *MetricsStorage) GetLatest() *models.SystemMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.metrics) == 0 {
		return nil
	}
	
	latest := s.metrics[len(s.metrics)-1]
	return &latest
}

// Clear removes all stored metrics
func (s *MetricsStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics = s.metrics[:0]
}

// Size returns the current number of stored metrics
func (s *MetricsStorage) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.metrics)
}