class SystemMonitor {
    constructor() {
        this.ws = null;
        this.cpuHistory = [];
        this.memoryHistory = [];
        this.maxHistoryPoints = 30;
        this.networkPrevious = {};
        this.lastUpdate = Date.now();
        
        this.initWebSocket();
        this.initCharts();
    }

    initWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            this.updateConnectionStatus(true);
            console.log('WebSocket connected');
        };
        
        this.ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.updateMetrics(data);
        };
        
        this.ws.onclose = () => {
            this.updateConnectionStatus(false);
            console.log('WebSocket disconnected');
            // Reconnect after 3 seconds
            setTimeout(() => this.initWebSocket(), 3000);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }

    initCharts() {
        this.cpuCanvas = document.getElementById('cpu-chart');
        this.cpuCtx = this.cpuCanvas.getContext('2d');
        
        this.memoryCanvas = document.getElementById('memory-chart');
        this.memoryCtx = this.memoryCanvas.getContext('2d');
    }

    updateConnectionStatus(connected) {
        const statusEl = document.getElementById('connection-status');
        if (connected) {
            statusEl.textContent = 'Connected';
            statusEl.className = 'connected';
        } else {
            statusEl.textContent = 'Disconnected';
            statusEl.className = 'disconnected';
        }
    }

    updateMetrics(data) {
        // Update last update time
        document.getElementById('last-update').textContent = 
            `Last update: ${new Date().toLocaleTimeString()}`;
        
        // Update system info
        this.updateSystemInfo(data.system);
        
        // Update CPU metrics
        this.updateCPU(data.cpu);
        
        // Update Memory metrics
        this.updateMemory(data.memory);
        
        // Update Disk metrics
        this.updateDisk(data.disk);
        
        // Update Network metrics
        this.updateNetwork(data.network);
        
        // Update charts
        this.updateCharts();
    }

    updateSystemInfo(system) {
        if (!system) return;
        
        document.getElementById('hostname').textContent = system.hostname || '-';
        document.getElementById('os').textContent = system.os || '-';
        document.getElementById('platform').textContent = system.platform || '-';
        document.getElementById('uptime').textContent = this.formatUptime(system.uptime || 0);
    }

    updateCPU(cpu) {
        if (!cpu) return;
        
        const percent = cpu.total_percent || 0;
        document.getElementById('cpu-percent').textContent = `${percent.toFixed(1)}%`;
        document.getElementById('cpu-cores').textContent = `${cpu.cores || 0} cores`;
        document.getElementById('cpu-bar').style.width = `${percent}%`;
        
        // Add to history
        this.cpuHistory.push(percent);
        if (this.cpuHistory.length > this.maxHistoryPoints) {
            this.cpuHistory.shift();
        }
        
        // Update CPU cores grid
        this.updateCPUCores(cpu.usage_percent);
    }

    updateCPUCores(coreUsages) {
        if (!coreUsages) return;
        
        const grid = document.getElementById('cpu-cores-grid');
        grid.innerHTML = '';
        
        coreUsages.forEach((usage, index) => {
            const coreDiv = document.createElement('div');
            coreDiv.className = 'cpu-core';
            coreDiv.innerHTML = `
                <div class="cpu-core-label">Core ${index}</div>
                <div class="cpu-core-value">${usage.toFixed(1)}%</div>
            `;
            grid.appendChild(coreDiv);
        });
    }

    updateMemory(memory) {
        if (!memory) return;
        
        const percent = memory.used_percent || 0;
        const usedGB = (memory.used / 1024 / 1024 / 1024).toFixed(1);
        const totalGB = (memory.total / 1024 / 1024 / 1024).toFixed(1);
        
        document.getElementById('memory-percent').textContent = `${percent.toFixed(1)}%`;
        document.getElementById('memory-details').textContent = `${usedGB} GB / ${totalGB} GB`;
        document.getElementById('memory-bar').style.width = `${percent}%`;
        
        // Add to history
        this.memoryHistory.push(percent);
        if (this.memoryHistory.length > this.maxHistoryPoints) {
            this.memoryHistory.shift();
        }
    }

    updateDisk(disks) {
        if (!disks) return;
        
        const diskList = document.getElementById('disk-list');
        diskList.innerHTML = '';
        
        disks.forEach(disk => {
            const usedGB = (disk.used / 1024 / 1024 / 1024).toFixed(1);
            const totalGB = (disk.total / 1024 / 1024 / 1024).toFixed(1);
            
            const diskItem = document.createElement('div');
            diskItem.className = 'disk-item';
            diskItem.innerHTML = `
                <div class="disk-header">
                    <span class="disk-name">${disk.mountpoint}</span>
                    <span>${disk.used_percent.toFixed(1)}%</span>
                </div>
                <div class="disk-usage">
                    <span>${usedGB} GB used</span>
                    <span>${totalGB} GB total</span>
                </div>
                <div class="disk-bar">
                    <div class="disk-bar-fill" style="width: ${disk.used_percent}%"></div>
                </div>
            `;
            diskList.appendChild(diskItem);
        });
    }

    updateNetwork(networks) {
        if (!networks) return;
        
        const networkList = document.getElementById('network-list');
        networkList.innerHTML = '';
        
        const currentTime = Date.now();
        const timeDiff = (currentTime - this.lastUpdate) / 1000; // seconds
        
        networks.forEach(network => {
            const prevNetwork = this.networkPrevious[network.name] || {};
            const sentRate = prevNetwork.bytes_sent ? 
                ((network.bytes_sent - prevNetwork.bytes_sent) / timeDiff) : 0;
            const recvRate = prevNetwork.bytes_recv ? 
                ((network.bytes_recv - prevNetwork.bytes_recv) / timeDiff) : 0;
            
            const networkItem = document.createElement('div');
            networkItem.className = 'network-item';
            networkItem.innerHTML = `
                <div class="network-header">
                    <span class="network-name">${network.name}</span>
                </div>
                <div class="network-grid">
                    <div class="network-stat">
                        <span class="label">Sent:</span>
                        <span class="value">${this.formatBytes(network.bytes_sent)}</span>
                    </div>
                    <div class="network-stat">
                        <span class="label">Received:</span>
                        <span class="value">${this.formatBytes(network.bytes_recv)}</span>
                    </div>
                    <div class="network-stat">
                        <span class="label">Send rate:</span>
                        <span class="value">${this.formatBytes(sentRate)}/s</span>
                    </div>
                    <div class="network-stat">
                        <span class="label">Recv rate:</span>
                        <span class="value">${this.formatBytes(recvRate)}/s</span>
                    </div>
                </div>
            `;
            networkList.appendChild(networkItem);
            
            // Store current values for next update
            this.networkPrevious[network.name] = {
                bytes_sent: network.bytes_sent,
                bytes_recv: network.bytes_recv
            };
        });
        
        this.lastUpdate = currentTime;
    }

    updateCharts() {
        this.drawChart(this.cpuCtx, this.cpuHistory, '#667eea');
        this.drawChart(this.memoryCtx, this.memoryHistory, '#764ba2');
    }

    drawChart(ctx, data, color) {
        const width = ctx.canvas.width;
        const height = ctx.canvas.height;
        
        // Clear canvas
        ctx.clearRect(0, 0, width, height);
        
        if (data.length < 2) return;
        
        // Draw grid lines
        ctx.strokeStyle = '#e5e7eb';
        ctx.lineWidth = 1;
        
        for (let i = 0; i <= 4; i++) {
            const y = (height / 4) * i;
            ctx.beginPath();
            ctx.moveTo(0, y);
            ctx.lineTo(width, y);
            ctx.stroke();
        }
        
        // Draw data line
        ctx.strokeStyle = color;
        ctx.lineWidth = 2;
        ctx.beginPath();
        
        const stepX = width / (this.maxHistoryPoints - 1);
        data.forEach((value, index) => {
            const x = index * stepX;
            const y = height - (value / 100) * height;
            
            if (index === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        });
        
        ctx.stroke();
        
        // Fill area under line
        ctx.lineTo(width, height);
        ctx.lineTo(0, height);
        ctx.closePath();
        
        const gradient = ctx.createLinearGradient(0, 0, 0, height);
        gradient.addColorStop(0, color + '40');
        gradient.addColorStop(1, color + '10');
        ctx.fillStyle = gradient;
        ctx.fill();
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    formatUptime(seconds) {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        const parts = [];
        if (days > 0) parts.push(`${days}d`);
        if (hours > 0) parts.push(`${hours}h`);
        if (minutes > 0) parts.push(`${minutes}m`);
        
        return parts.length > 0 ? parts.join(' ') : '0m';
    }
}

// Initialize monitor when page loads
document.addEventListener('DOMContentLoaded', () => {
    new SystemMonitor();
});