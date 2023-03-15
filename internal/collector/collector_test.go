package collector

import (
	"runtime"
	"testing"
)

func TestNewCollector(t *testing.T) {
	c := NewCollector()
	
	if c == nil {
		t.Fatal("Expected collector to be created")
	}
	
	if c.lastNetworkStats == nil {
		t.Error("Expected lastNetworkStats map to be initialized")
	}
}

func TestCollect(t *testing.T) {
	c := NewCollector()
	
	metrics, err := c.Collect()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Check timestamp
	if metrics.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}
	
	// Check CPU metrics
	if metrics.CPU.Cores == 0 {
		t.Error("Expected CPU cores to be greater than 0")
	}
	
	if len(metrics.CPU.UsagePercent) == 0 {
		t.Error("Expected CPU usage percent to have values")
	}
	
	// Check Memory metrics
	if metrics.Memory.Total == 0 {
		t.Error("Expected total memory to be greater than 0")
	}
	
	// Check System info
	if metrics.System.Hostname == "" {
		t.Error("Expected hostname to be set")
	}
	
	if metrics.System.OS == "" {
		t.Error("Expected OS to be set")
	}
	
	if metrics.System.Platform == "" {
		t.Error("Expected platform to be set")
	}
}

func TestCollectCPU(t *testing.T) {
	c := NewCollector()
	
	cpu, err := c.collectCPU()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if cpu.Cores == 0 {
		t.Error("Expected CPU cores to be greater than 0")
	}
	
	if len(cpu.UsagePercent) == 0 {
		t.Error("Expected CPU usage percent to have values")
	}
	
	if cpu.TotalPercent < 0 || cpu.TotalPercent > 100 {
		t.Errorf("Expected CPU total percent to be between 0 and 100, got %f", cpu.TotalPercent)
	}
	
	// Check load average on Unix-like systems
	if runtime.GOOS != "windows" {
		if len(cpu.LoadAvg) != 3 {
			t.Errorf("Expected 3 load average values, got %d", len(cpu.LoadAvg))
		}
	}
}

func TestCollectMemory(t *testing.T) {
	c := NewCollector()
	
	mem, err := c.collectMemory()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if mem.Total == 0 {
		t.Error("Expected total memory to be greater than 0")
	}
	
	if mem.Total < mem.Used {
		t.Error("Expected total memory to be greater than or equal to used memory")
	}
	
	if mem.UsedPercent < 0 || mem.UsedPercent > 100 {
		t.Errorf("Expected memory used percent to be between 0 and 100, got %f", mem.UsedPercent)
	}
}

func TestCollectDisk(t *testing.T) {
	c := NewCollector()
	
	disks, err := c.collectDisk()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(disks) == 0 {
		t.Skip("No disk partitions found")
	}
	
	for _, disk := range disks {
		if disk.Total == 0 {
			t.Error("Expected disk total to be greater than 0")
		}
		
		if disk.Device == "" {
			t.Error("Expected disk device to be set")
		}
		
		if disk.Mountpoint == "" {
			t.Error("Expected disk mountpoint to be set")
		}
		
		if disk.UsedPercent < 0 || disk.UsedPercent > 100 {
			t.Errorf("Expected disk used percent to be between 0 and 100, got %f", disk.UsedPercent)
		}
	}
}

func TestCollectNetwork(t *testing.T) {
	c := NewCollector()
	
	networks, err := c.collectNetwork()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Network interfaces might not be available in test environment
	if len(networks) > 0 {
		for _, net := range networks {
			if net.Name == "" {
				t.Error("Expected network interface name to be set")
			}
		}
	}
}

func TestCollectSystem(t *testing.T) {
	c := NewCollector()
	
	sys, err := c.collectSystem()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if sys.Hostname == "" {
		t.Error("Expected hostname to be set")
	}
	
	if sys.OS == "" {
		t.Error("Expected OS to be set")
	}
	
	if sys.Platform == "" {
		t.Error("Expected platform to be set")
	}
	
	if sys.BootTime == 0 {
		t.Error("Expected boot time to be set")
	}
	
	if sys.Uptime == 0 {
		t.Error("Expected uptime to be greater than 0")
	}
}

func TestShouldSkipFilesystem(t *testing.T) {
	tests := []struct {
		fstype   string
		expected bool
	}{
		{"ext4", false},
		{"ntfs", false},
		{"apfs", false},
		{"tmpfs", true},
		{"devfs", true},
		{"proc", true},
		{"sysfs", true},
		{"overlay", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.fstype, func(t *testing.T) {
			result := shouldSkipFilesystem(tt.fstype)
			if result != tt.expected {
				t.Errorf("shouldSkipFilesystem(%s) = %v, want %v", tt.fstype, result, tt.expected)
			}
		})
	}
}