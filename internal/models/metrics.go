package models

import "time"

// SystemMetrics represents all system metrics collected at a point in time
type SystemMetrics struct {
	Timestamp   time.Time        `json:"timestamp"`
	CPU         CPUMetrics       `json:"cpu"`
	Memory      MemoryMetrics    `json:"memory"`
	Disk        []DiskMetrics    `json:"disk"`
	Network     []NetworkMetrics `json:"network"`
	System      SystemInfo       `json:"system"`
	Temperature []TempMetrics    `json:"temperature,omitempty"`
}

// CPUMetrics represents CPU usage information
type CPUMetrics struct {
	UsagePercent []float64 `json:"usage_percent"`
	TotalPercent float64   `json:"total_percent"`
	Cores        int       `json:"cores"`
	LoadAvg      []float64 `json:"load_avg,omitempty"`
}

// MemoryMetrics represents memory usage information
type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	Available   uint64  `json:"available"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapFree    uint64  `json:"swap_free"`
	SwapPercent float64 `json:"swap_percent"`
}

// DiskMetrics represents disk usage information
type DiskMetrics struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

// NetworkMetrics represents network interface statistics
type NetworkMetrics struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	Errin       uint64 `json:"errin"`
	Errout      uint64 `json:"errout"`
	Dropin      uint64 `json:"dropin"`
	Dropout     uint64 `json:"dropout"`
}

// SystemInfo represents general system information
type SystemInfo struct {
	Hostname        string `json:"hostname"`
	OS              string `json:"os"`
	Platform        string `json:"platform"`
	PlatformVersion string `json:"platform_version"`
	KernelVersion   string `json:"kernel_version"`
	Uptime          uint64 `json:"uptime"`
	BootTime        uint64 `json:"boot_time"`
	Processes       uint64 `json:"processes"`
}

// TempMetrics represents temperature sensor readings
type TempMetrics struct {
	SensorKey   string  `json:"sensor_key"`
	Temperature float64 `json:"temperature"`
	Label       string  `json:"label,omitempty"`
}