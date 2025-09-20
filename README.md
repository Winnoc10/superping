# SuperPing - Network Monitor

A clean, travel-friendly CLI tool to monitor network connectivity while on trains, planes, or with unreliable WiFi.

## Features

- **Real-time monitoring** of multiple network targets
- **Bandwidth micro-tests** - Quick throughput measurements without consuming much data
- **Jitter detection** - Measures connection stability for video calls
- **Quality scoring** - 0-100 score combining latency, jitter, bandwidth, and reliability
- **Visual trend indicators** showing connection history
- **Fullscreen TUI** with clean, non-technical interface
- **Auto-refresh** every 2 seconds
- **Multiple check types**: ping (TCP), HTTP, DNS resolution, and bandwidth tests

## Quick Start

```bash
# Build the tool
./build.sh

# Run SuperPing
./superping
```

Or directly with Go:
```bash
go run main.go
```

## What it monitors

### 🌐 DNS & Connectivity
- Google DNS (8.8.8.8) - Basic connectivity test
- Cloudflare (1.1.1.1) - Fast DNS provider
- Quad9 (9.9.9.9) - Privacy-focused DNS (Europe-friendly)
- OpenDNS (208.67.222.222) - Alternative DNS service
- DNS Resolution - Domain name resolution test

### 🌍 Website Access
- Google (HTTPS) - Web browsing capability
- GitHub (HTTPS) - Development platform access
- BBC (HTTPS) - Major European news site
- Wikipedia (HTTPS) - Global knowledge platform
- Stack Overflow (HTTPS) - Developer community

### ⚡ Speed Test
- Bandwidth Test - Throughput measurement (10KB download)

## Interface

The tool shows:
- 🟢 **GOOD** - Connection working well
- 🟡 **SLOW** - Connection but with higher latency
- 🟠 **TIMEOUT** - Connection timed out
- 🔴 **ERROR** - Connection failed

### New Metrics Displayed
- **Latency ± Jitter** - Response time and stability (e.g., "45ms ±12ms")
- **Jitter Bar** - Visual stability indicator below latency:
  - 🟢 **Green** = Very stable (< 5ms jitter)
  - 🟡 **Yellow** = Moderate stability (5-20ms jitter)
  - 🟠 **Orange** = Unstable (20-50ms jitter)
  - 🔴 **Red** = Very unstable (>50ms jitter)
- **Bandwidth** - Download speed for the speed test target (e.g., "8.3 Mbps")
- **Quality Score** - Overall connection quality 0-100 (e.g., "Q:85")

Each target shows a mini trend line showing recent connection history.

## Controls

- `q` or `Ctrl+C` to quit

## Travel Tips

- Leave it running to monitor connection quality over time
- The trend lines help identify patterns in connectivity
- Overall network health shown at the bottom

Perfect for understanding your connection quality while working remotely!