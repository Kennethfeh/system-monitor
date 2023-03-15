package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

//go:embed static/*
var staticFiles embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type SystemMetrics struct {
	Timestamp   int64            `json:"timestamp"`
	CPU         CPUMetrics       `json:"cpu"`
	Memory      MemoryMetrics    `json:"memory"`
	Disk        []DiskMetrics    `json:"disk"`
	Network     []NetworkMetrics `json:"network"`
	System      SystemInfo       `json:"system"`
}

type CPUMetrics struct {
	UsagePercent []float64 `json:"usage_percent"`
	TotalPercent float64   `json:"total_percent"`
	Cores        int       `json:"cores"`
}

type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskMetrics struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkMetrics struct {
	Name        string `json:"name"`
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

type SystemInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Uptime   uint64 `json:"uptime"`
}

func collectMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now().Unix(),
	}

	// CPU metrics
	cpuPercent, err := cpu.Percent(time.Second, true)
	if err == nil {
		metrics.CPU.UsagePercent = cpuPercent
		total := 0.0
		for _, p := range cpuPercent {
			total += p
		}
		if len(cpuPercent) > 0 {
			metrics.CPU.TotalPercent = total / float64(len(cpuPercent))
		}
	}
	
	cpuCount, err := cpu.Counts(true)
	if err == nil {
		metrics.CPU.Cores = cpuCount
	}

	// Memory metrics
	vmStat, err := mem.VirtualMemory()
	if err == nil {
		metrics.Memory = MemoryMetrics{
			Total:       vmStat.Total,
			Used:        vmStat.Used,
			Free:        vmStat.Free,
			UsedPercent: vmStat.UsedPercent,
		}
	}

	// Disk metrics
	partitions, err := disk.Partitions(false)
	if err == nil {
		for _, partition := range partitions {
			usage, err := disk.Usage(partition.Mountpoint)
			if err == nil && usage.Total > 0 {
				metrics.Disk = append(metrics.Disk, DiskMetrics{
					Device:      partition.Device,
					Mountpoint:  partition.Mountpoint,
					Total:       usage.Total,
					Used:        usage.Used,
					Free:        usage.Free,
					UsedPercent: usage.UsedPercent,
				})
			}
		}
	}

	// Network metrics
	netIO, err := net.IOCounters(true)
	if err == nil {
		for _, io := range netIO {
			if io.BytesSent > 0 || io.BytesRecv > 0 {
				metrics.Network = append(metrics.Network, NetworkMetrics{
					Name:        io.Name,
					BytesSent:   io.BytesSent,
					BytesRecv:   io.BytesRecv,
					PacketsSent: io.PacketsSent,
					PacketsRecv: io.PacketsRecv,
				})
			}
		}
	}

	// System info
	hostInfo, err := host.Info()
	if err == nil {
		metrics.System = SystemInfo{
			Hostname: hostInfo.Hostname,
			OS:       hostInfo.OS,
			Platform: hostInfo.Platform,
			Uptime:   hostInfo.Uptime,
		}
	}

	return metrics, nil
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics, err := collectMetrics()
			if err != nil {
				log.Printf("Failed to collect metrics: %v", err)
				continue
			}

			if err := conn.WriteJSON(metrics); err != nil {
				log.Printf("Failed to write metrics: %v", err)
				return
			}
		}
	}
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	metrics, err := collectMetrics()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func main() {
	// Serve static files
	http.Handle("/", http.FileServer(http.FS(staticFiles)))
	
	// API endpoints
	http.HandleFunc("/api/metrics", handleAPI)
	http.HandleFunc("/ws", handleWebSocket)

	port := ":8080"
	fmt.Printf("System Monitor started on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}