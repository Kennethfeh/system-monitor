# System Monitor

A real-time system monitoring tool built with Go, featuring a web interface for tracking CPU, memory, disk, and network metrics.

## Features

- **Real-time Monitoring**: Live updates via WebSocket connection
- **CPU Metrics**: Overall usage, per-core usage, and historical charts
- **Memory Tracking**: RAM usage with visual progress bars and charts
- **Disk Usage**: Monitor multiple drives and partitions
- **Network Statistics**: Track network interface traffic and rates
- **System Information**: Display hostname, OS, platform, and uptime
- **Responsive Web UI**: Clean, modern interface with live charts

## Prerequisites

- Go 1.20 or higher
- Web browser with WebSocket support

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd system-monitor
```

2. Download dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build
```

## Usage

1. Run the built binary:
```bash
./system-monitor
```

2. Open your web browser and navigate to:
```
http://localhost:8080
```

The web interface will automatically connect via WebSocket and begin displaying real-time system metrics.

## API Endpoints

- `/` - Web interface
- `/api/metrics` - REST endpoint for current metrics (JSON)
- `/ws` - WebSocket endpoint for real-time updates

## Project Structure

```
system-monitor/
├── main.go           # Main application with server and metrics collection
├── go.mod           # Go module definition
├── go.sum           # Dependency checksums
├── static/          # Web interface files
│   ├── index.html   # Main HTML page
│   ├── style.css    # Styling
│   └── app.js       # JavaScript client
└── README.md        # This file
```

## Dependencies

- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [gopsutil](https://github.com/shirou/gopsutil) - System and process utilities

## Technical Details

- Built with Go's embed feature for serving static files
- WebSocket updates every 2 seconds
- Historical data maintained for charts (last 30 data points)
- Network rate calculations based on delta between updates
- Cross-platform support via gopsutil library

## Browser Compatibility

Works with all modern browsers supporting:
- WebSocket API
- Canvas API for charts
- ES6 JavaScript features

## License

MIT License

## Author

Kenneth Feh (2023)