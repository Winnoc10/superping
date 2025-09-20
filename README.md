# SuperPing - Network Monitor

🚄 A clean, travel-friendly CLI tool to monitor network connectivity while on trains, planes, or with unreliable WiFi.

## Preview

```
⚡ SuperPing - Network Monitor
Last check: 15:42:33

🌐 DNS
┌─ Google DNS ────┐ ┌─ Cloudflare ───┐ ┌─ Quad9 ────────┐ ┌─ OpenDNS ──────┐ ┌─ DNS Resolution ┐
│ GOOD            │ │ GOOD           │ │ GOOD           │ │ SLOW           │ │ GOOD            │
│ 23ms ±8ms       │ │ 18ms ±5ms      │ │ 45ms ±12ms     │ │ 156ms ±89ms    │ │ 34ms ±15ms      │
│ ▊▊▁▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▊▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▊▊▊▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▊▊▊▊▊▊▊▊▊▊▊▊▊▊▊ │ │ ▊▊▊▁▁▁▁▁▁▁▁▁▁▁▁ │
│ Q:95            │ │ Q:98           │ │ Q:88           │ │ Q:45           │ │ Q:82            │
│ ▁▁▁▃▁▁▁▁▁▃▁▁▁▁▁ │ │ ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▁▁▃▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▃▅▅▃▃▅▁▁▁▁▁▁▁▁▁ │ │ ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │
└─────────────────┘ └────────────────┘ └────────────────┘ └────────────────┘ └─────────────────┘

🌍 Web
┌─ Google ────────┐ ┌─ GitHub ───────┐ ┌─ BBC ──────────┐ ┌─ Wikipedia ─────┐ ┌─ Stack Overflow ┐
│ GOOD            │ │ GOOD           │ │ SLOW           │ │ GOOD           │ │ GOOD            │
│ 45ms ±23ms      │ │ 67ms ±34ms     │ │ 234ms ±78ms    │ │ 89ms ±45ms     │ │ 78ms ±28ms      │
│ ▊▊▊▊▊▁▁▁▁▁▁▁▁▁▁ │ │ ▊▊▊▊▊▊▊▁▁▁▁▁▁▁▁ │ │ ▊▊▊▊▊▊▊▊▊▊▊▊▊▊▊ │ │ ▊▊▊▊▊▊▊▊▁▁▁▁▁▁▁ │ │ ▊▊▊▊▊▊▁▁▁▁▁▁▁▁▁ │
│ Q:78            │ │ Q:72           │ │ Q:35           │ │ Q:65           │ │ Q:69            │
│ ▁▁▁▁▁▃▁▁▁▁▁▁▁▁▁ │ │ ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▃▅▅▃▃▅▁▁▁▁▁▁▁▁▁ │ │ ▁▁▃▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │
└─────────────────┘ └────────────────┘ └────────────────┘ └────────────────┘ └─────────────────┘

⚡ Speed
┌─ Download Test ─┐ ┌─ Upload Test ───┐
│ SLOW            │ │ TIMEOUT        │
│ 156ms ±45ms     │ │ 892ms ±234ms   │
│ ▊▊▊▊▊▊▊▁▁▁▁▁▁▁▁ │ │ ▊▊▊▊▊▊▊▊▊▊▊▊▊▊▊ │
│ ↓1.2 Mbps Q:42  │ │ ↑0.1 Mbps Q:15 │
│ ▃▅▅▃▃▁▁▁▁▁▁▁▁▁▁ │ │ ▅█▅▃▃▅▁▁▁▁▁▁▁▁▁ │
└─────────────────┘ └────────────────┘

🔗 Connectivity
┌─ Captive Portal ┐ ┌─ IPv6 Support ─┐ ┌─ Route Hops ────┐
│ GOOD            │ │ SLOW           │ │ GOOD            │
│ 67ms ±12ms      │ │ 89ms ±23ms     │ │ 45ms ±8ms       │
│ ▊▊▊▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▊▊▊▊▊▁▁▁▁▁▁▁▁▁▁ │ │ ▊▊▁▁▁▁▁▁▁▁▁▁▁▁▁ │
│ No portal Q:82  │ │ IPv4 only Q:68 │ │ 12 hops Q:85    │
│ ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▁▁▃▁▁▁▁▁▁▁▁▁▁▁▁ │ │ ▁▁▁▁▁▁▁▁▁▁▁▁▁▁▁ │
└─────────────────┘ └────────────────┘ └─────────────────┘

Network Health: GOOD (11/16 targets)
Press 'q' or Ctrl+C to quit
```

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

### ⚡ Speed Tests
- Download Test - Throughput measurement (10KB download)
- Upload Test - Upload speed measurement (5KB upload)

### 🔗 Connectivity Tests
- Captive Portal - Detects hotel/airport login requirements
- IPv6 Support - Tests IPv6 availability vs IPv4-only networks
- Route Hops - Network path length via traceroute

## Interface

The tool shows:
- 🟢 **GOOD** - Connection working well
- 🟡 **SLOW** - Connection but with higher latency
- 🟠 **TIMEOUT** - Connection timed out
- 🔴 **ERROR** - Connection failed

### Metrics Displayed
- **Latency ± Jitter** - Response time and stability (e.g., "45ms ±12ms")
- **Jitter Bar** - Visual stability indicator below latency:
  - 🟢 **Green** = Very stable (< 5ms jitter)
  - 🟡 **Yellow** = Moderate stability (5-20ms jitter)
  - 🟠 **Orange** = Unstable (20-50ms jitter)
  - 🔴 **Red** = Very unstable (>50ms jitter)
- **Download/Upload Speed** - Real bandwidth measurements (e.g., "↓2.3 Mbps", "↑0.8 Mbps")
- **Connectivity Status** - Special indicators (e.g., "No portal", "IPv6 available", "12 hops")
- **Quality Score** - Overall connection quality 0-100 (e.g., "Q:85")

Each target shows a mini trend line showing recent connection history.

## Controls

- `q` or `Ctrl+C` to quit

## Travel Tips

- Leave it running to monitor connection quality over time
- The trend lines help identify patterns in connectivity
- Overall network health shown at the bottom

Perfect for understanding your connection quality while working remotely!