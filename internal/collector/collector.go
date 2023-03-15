package collector

import (
	"runtime"
	"time"

	"github.com/kennethfeh/system-monitor/internal/models"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/shirou/gopsutil/v3/process"
)

// Collector handles system metrics collection
type Collector struct {
	lastNetworkStats map[string]models.NetworkMetrics
	lastCollectTime  time.Time
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	return &Collector{
		lastNetworkStats: make(map[string]models.NetworkMetrics),
		lastCollectTime:  time.Now(),
	}
}

// Collect gathers all system metrics
func (c *Collector) Collect() (models.SystemMetrics, error) {
	metrics := models.SystemMetrics{
		Timestamp: time.Now(),
	}

	// Collect CPU metrics
	if cpuMetrics, err := c.collectCPU(); err == nil {
		metrics.CPU = cpuMetrics
	}

	// Collect Memory metrics
	if memMetrics, err := c.collectMemory(); err == nil {
		metrics.Memory = memMetrics
	}

	// Collect Disk metrics
	if diskMetrics, err := c.collectDisk(); err == nil {
		metrics.Disk = diskMetrics
	}

	// Collect Network metrics
	if netMetrics, err := c.collectNetwork(); err == nil {
		metrics.Network = netMetrics
	}

	// Collect System info
	if sysInfo, err := c.collectSystem(); err == nil {
		metrics.System = sysInfo
	}

	// Collect Temperature (if available)
	if tempMetrics, err := c.collectTemperature(); err == nil && len(tempMetrics) > 0 {
		metrics.Temperature = tempMetrics
	}

	c.lastCollectTime = time.Now()
	return metrics, nil
}

func (c *Collector) collectCPU() (models.CPUMetrics, error) {
	cpuMetrics := models.CPUMetrics{}

	// Get CPU usage per core
	cpuPercent, err := cpu.Percent(time.Second, true)
	if err != nil {
		return cpuMetrics, err
	}
	cpuMetrics.UsagePercent = cpuPercent

	// Calculate total CPU usage
	if len(cpuPercent) > 0 {
		total := 0.0
		for _, p := range cpuPercent {
			total += p
		}
		cpuMetrics.TotalPercent = total / float64(len(cpuPercent))
	}

	// Get CPU core count
	cpuCount, err := cpu.Counts(true)
	if err == nil {
		cpuMetrics.Cores = cpuCount
	}

	// Get load average (Unix-like systems)
	if runtime.GOOS != "windows" {
		if loadAvg, err := load.Avg(); err == nil {
			cpuMetrics.LoadAvg = []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}
		}
	}

	return cpuMetrics, nil
}

func (c *Collector) collectMemory() (models.MemoryMetrics, error) {
	memMetrics := models.MemoryMetrics{}

	// Virtual memory
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		return memMetrics, err
	}

	memMetrics.Total = vmStat.Total
	memMetrics.Used = vmStat.Used
	memMetrics.Free = vmStat.Free
	memMetrics.Available = vmStat.Available
	memMetrics.UsedPercent = vmStat.UsedPercent

	// Swap memory
	swapStat, err := mem.SwapMemory()
	if err == nil {
		memMetrics.SwapTotal = swapStat.Total
		memMetrics.SwapUsed = swapStat.Used
		memMetrics.SwapFree = swapStat.Free
		memMetrics.SwapPercent = swapStat.UsedPercent
	}

	return memMetrics, nil
}

func (c *Collector) collectDisk() ([]models.DiskMetrics, error) {
	var diskMetrics []models.DiskMetrics

	partitions, err := disk.Partitions(false)
	if err != nil {
		return diskMetrics, err
	}

	for _, partition := range partitions {
		// Skip special filesystems
		if shouldSkipFilesystem(partition.Fstype) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil || usage.Total == 0 {
			continue
		}

		diskMetrics = append(diskMetrics, models.DiskMetrics{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Used:        usage.Used,
			Free:        usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}

	return diskMetrics, nil
}

func (c *Collector) collectNetwork() ([]models.NetworkMetrics, error) {
	var netMetrics []models.NetworkMetrics

	netIO, err := net.IOCounters(true)
	if err != nil {
		return netMetrics, err
	}

	for _, io := range netIO {
		// Skip loopback and inactive interfaces
		if io.Name == "lo" || io.Name == "lo0" || (io.BytesSent == 0 && io.BytesRecv == 0) {
			continue
		}

		metric := models.NetworkMetrics{
			Name:        io.Name,
			BytesSent:   io.BytesSent,
			BytesRecv:   io.BytesRecv,
			PacketsSent: io.PacketsSent,
			PacketsRecv: io.PacketsRecv,
			Errin:       io.Errin,
			Errout:      io.Errout,
			Dropin:      io.Dropin,
			Dropout:     io.Dropout,
		}

		netMetrics = append(netMetrics, metric)
	}

	return netMetrics, nil
}

func (c *Collector) collectSystem() (models.SystemInfo, error) {
	sysInfo := models.SystemInfo{}

	hostInfo, err := host.Info()
	if err != nil {
		return sysInfo, err
	}

	sysInfo.Hostname = hostInfo.Hostname
	sysInfo.OS = hostInfo.OS
	sysInfo.Platform = hostInfo.Platform
	sysInfo.PlatformVersion = hostInfo.PlatformVersion
	sysInfo.KernelVersion = hostInfo.KernelVersion
	sysInfo.Uptime = hostInfo.Uptime
	sysInfo.BootTime = hostInfo.BootTime

	// Get process count
	processes, err := process.Processes()
	if err == nil {
		sysInfo.Processes = uint64(len(processes))
	}

	return sysInfo, nil
}

func (c *Collector) collectTemperature() ([]models.TempMetrics, error) {
	var tempMetrics []models.TempMetrics

	// Temperature sensors are platform-specific
	// This is a simplified version - real implementation would need platform-specific code
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return tempMetrics, err
	}

	for _, temp := range temps {
		if temp.Temperature > 0 {
			tempMetrics = append(tempMetrics, models.TempMetrics{
				SensorKey:   temp.SensorKey,
				Temperature: temp.Temperature,
				Label:       temp.SensorKey,
			})
		}
	}

	return tempMetrics, nil
}

func shouldSkipFilesystem(fstype string) bool {
	skipList := []string{
		"devfs", "devtmpfs", "tmpfs", "sysfs", "proc",
		"cgroup", "cgroup2", "cpuset", "configfs", "debugfs",
		"tracefs", "securityfs", "pstore", "autofs", "mqueue",
		"hugetlbfs", "fusectl", "rpc_pipefs", "overlay", "squashfs",
	}

	for _, skip := range skipList {
		if fstype == skip {
			return true
		}
	}
	return false
}