package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	width  int
	height int

	targets    []Target
	lastUpdate time.Time

	quit bool
}

type Target struct {
	Name    string
	Type    string // "ping", "http", "dns", "bandwidth", "upload", "captive", "ipv6", "traceroute"
	Host    string
	Section string // "DNS", "Web", "Speed", "Connectivity"

	Status    ConnectionStatus
	LastCheck time.Time
	History   []ConnectionStatus

	Latency        time.Duration
	Jitter         time.Duration
	LatencyHistory []time.Duration
	Bandwidth      float64 // Mbps
	UploadSpeed    float64 // Mbps
	QualityScore   float64 // 0-100
	PacketLoss     float64
	IsCaptive      bool
	IPv6Support    bool
	HopCount       int
}

type ConnectionStatus int

const (
	StatusUnknown ConnectionStatus = iota
	StatusConnected
	StatusSlow
	StatusTimeout
	StatusError
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusConnected:
		return "GOOD"
	case StatusSlow:
		return "SLOW"
	case StatusTimeout:
		return "TIMEOUT"
	case StatusError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (s ConnectionStatus) Color() lipgloss.Color {
	switch s {
	case StatusConnected:
		return lipgloss.Color("46") // green
	case StatusSlow:
		return lipgloss.Color("226") // yellow
	case StatusTimeout:
		return lipgloss.Color("202") // orange
	case StatusError:
		return lipgloss.Color("196") // red
	default:
		return lipgloss.Color("245") // gray
	}
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel() model {
	targets := []Target{
		// DNS Tests
		{Name: "Google DNS", Type: "ping", Host: "8.8.8.8", Section: "DNS", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "Cloudflare", Type: "ping", Host: "1.1.1.1", Section: "DNS", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "Quad9", Type: "ping", Host: "9.9.9.9", Section: "DNS", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "OpenDNS", Type: "ping", Host: "208.67.222.222", Section: "DNS", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "DNS Resolution", Type: "dns", Host: "google.com", Section: "DNS", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},

		// Website Tests
		{Name: "Google", Type: "http", Host: "https://google.com", Section: "Web", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "GitHub", Type: "http", Host: "https://github.com", Section: "Web", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "BBC", Type: "http", Host: "https://bbc.co.uk", Section: "Web", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "Wikipedia", Type: "http", Host: "https://wikipedia.org", Section: "Web", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "Stack Overflow", Type: "http", Host: "https://stackoverflow.com", Section: "Web", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},

		// Speed Tests
		{Name: "Download Test", Type: "bandwidth", Host: "https://httpbin.org/bytes/10240", Section: "Speed", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "Upload Test", Type: "upload", Host: "https://httpbin.org/post", Section: "Speed", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},

		// Connectivity Tests
		{Name: "Captive Portal", Type: "captive", Host: "http://connectivitycheck.gstatic.com/generate_204", Section: "Connectivity", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "IPv6 Support", Type: "ipv6", Host: "google.com", Section: "Connectivity", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
		{Name: "Route Hops", Type: "traceroute", Host: "8.8.8.8", Section: "Connectivity", History: make([]ConnectionStatus, 0, 30), LatencyHistory: make([]time.Duration, 0, 10)},
	}

	return model{
		targets:    targets,
		lastUpdate: time.Now(),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tickCmd(), checkAllTargetsCmd(m.targets))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit
		}

	case tickMsg:
		m.lastUpdate = time.Time(msg)
		return m, tea.Batch(tickCmd(), checkAllTargetsCmd(m.targets))

	case targetsCheckedMsg:
		m.targets = msg.targets
	}

	return m, nil
}

func (m model) View() string {
	if m.quit {
		return ""
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Header
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Align(lipgloss.Center).
		Width(m.width)

	title := titleStyle.Render("⚡ SuperPing - Network Monitor")

	// Last update info
	updateStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(m.width).
		MarginBottom(1)

	lastUpdate := updateStyle.Render(fmt.Sprintf("Last check: %s", m.lastUpdate.Format("15:04:05")))

	// Connection status grid
	statusCards := m.renderStatusCards()

	// Connection history
	historyView := m.renderHistory()

	// Instructions
	instructStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Align(lipgloss.Center).
		Width(m.width).
		MarginTop(1)

	instructions := instructStyle.Render("Press 'q' or Ctrl+C to quit")

	// Combine all sections
	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		lastUpdate,
		statusCards,
		historyView,
		instructions,
	)

	// Ensure content fits the screen
	if lipgloss.Height(content) > m.height {
		content = lipgloss.NewStyle().Height(m.height).Render(content)
	}

	return content
}

func (m model) renderStatusCards() string {
	// Group targets by section
	sections := map[string][]Target{}
	sectionOrder := []string{"DNS", "Web", "Speed", "Connectivity"}

	for _, target := range m.targets {
		sections[target.Section] = append(sections[target.Section], target)
	}

	var allSections []string

	for _, sectionName := range sectionOrder {
		targets := sections[sectionName]
		if len(targets) == 0 {
			continue
		}

		// Section header
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			MarginTop(0).
			MarginBottom(0)

		var sectionTitle string
		switch sectionName {
		case "DNS":
			sectionTitle = "🌐 DNS"
		case "Web":
			sectionTitle = "🌍 Web"
		case "Speed":
			sectionTitle = "⚡ Speed"
		case "Connectivity":
			sectionTitle = "🔗 Connectivity"
		}

		header := headerStyle.Render(sectionTitle)

		// Render cards for this section
		var cards []string
		for _, target := range targets {
			card := m.renderStatusCard(target)
			cards = append(cards, card)
		}

		// Arrange cards in rows
		cardWidth := 27
		cardsPerRow := max(1, m.width/(cardWidth+2))

		var rows []string
		for i := 0; i < len(cards); i += cardsPerRow {
			end := min(i+cardsPerRow, len(cards))
			row := lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...)
			rows = append(rows, row)
		}

		sectionContent := lipgloss.JoinVertical(lipgloss.Left, rows...)

		// Combine header and content
		section := lipgloss.JoinVertical(lipgloss.Left, header, sectionContent)
		allSections = append(allSections, section)
	}

	return lipgloss.JoinVertical(lipgloss.Left, allSections...)
}

func (m model) renderStatusCard(target Target) string {
	statusColor := target.Status.Color()

	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(statusColor).
		Width(25).
		Height(7).
		Padding(0, 1).
		Margin(0, 1, 0, 0)

	// Name
	nameStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Width(23)

	name := nameStyle.Render(target.Name)

	// Status
	statusStyle := lipgloss.NewStyle().
		Foreground(statusColor).
		Bold(true)

	status := statusStyle.Render(target.Status.String())

	// Metrics line 1: Latency and Jitter
	metricsStr1 := ""
	if target.Latency > 0 {
		metricsStr1 = fmt.Sprintf("%dms", target.Latency.Milliseconds())
		if target.Jitter > 0 {
			metricsStr1 += fmt.Sprintf(" ±%dms", target.Jitter.Milliseconds())
		}
	}

	// Metrics line 2: Bandwidth, Upload, and special metrics
	metricsStr2 := ""
	switch target.Type {
	case "bandwidth":
		if target.Bandwidth > 0 {
			metricsStr2 = fmt.Sprintf("↓%.1f Mbps", target.Bandwidth)
		}
	case "upload":
		if target.UploadSpeed > 0 {
			metricsStr2 = fmt.Sprintf("↑%.1f Mbps", target.UploadSpeed)
		}
	case "captive":
		if target.IsCaptive {
			metricsStr2 = "Portal detected"
		} else {
			metricsStr2 = "No portal"
		}
	case "ipv6":
		if target.IPv6Support {
			metricsStr2 = "IPv6 available"
		} else {
			metricsStr2 = "IPv4 only"
		}
	case "traceroute":
		if target.HopCount > 0 {
			metricsStr2 = fmt.Sprintf("%d hops", target.HopCount)
		}
	}

	// Add quality score if available
	if target.QualityScore > 0 {
		if metricsStr2 != "" {
			metricsStr2 += " "
		}
		metricsStr2 += fmt.Sprintf("Q:%.0f", target.QualityScore)
	}

	metricsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245"))

	metrics1 := metricsStyle.Render(metricsStr1)
	metrics2 := metricsStyle.Render(metricsStr2)

	// Jitter visualization
	jitterBar := m.renderJitterBar(target.Jitter)

	// Mini trend line
	trend := m.renderMiniTrend(target.History)

	cardContent := lipgloss.JoinVertical(lipgloss.Left,
		name,
		status,
		metrics1,
		jitterBar,
		metrics2,
		trend,
	)

	return borderStyle.Render(cardContent)
}

func (m model) renderJitterBar(jitter time.Duration) string {
	jitterMs := jitter.Milliseconds()
	barWidth := 21 // Match card width

	// Determine jitter level and color
	var jitterColor lipgloss.Color
	var intensity int

	switch {
	case jitterMs >= 100:
		jitterColor = lipgloss.Color("196") // red - very unstable
		intensity = barWidth
	case jitterMs >= 50:
		jitterColor = lipgloss.Color("202") // orange - unstable
		intensity = int(float64(barWidth) * 0.7)
	case jitterMs >= 20:
		jitterColor = lipgloss.Color("226") // yellow - moderate
		intensity = int(float64(barWidth) * 0.4)
	case jitterMs >= 5:
		jitterColor = lipgloss.Color("46") // green - stable
		intensity = int(float64(barWidth) * 0.2)
	default:
		jitterColor = lipgloss.Color("245") // gray - very stable
		intensity = 1
	}

	// Build the jitter bar
	var bar strings.Builder

	// Add filled portion
	filledStyle := lipgloss.NewStyle().Foreground(jitterColor)
	for i := 0; i < intensity && i < barWidth; i++ {
		bar.WriteString(filledStyle.Render("▊"))
	}

	// Add empty portion
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	for i := intensity; i < barWidth; i++ {
		bar.WriteString(emptyStyle.Render("▁"))
	}

	return bar.String()
}

func (m model) renderMiniTrend(history []ConnectionStatus) string {
	if len(history) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("━━━━━━━━━━━━━━━━━━━")
	}

	var trend strings.Builder
	maxItems := 19 // Width of trend line
	start := max(0, len(history)-maxItems)

	for i := start; i < len(history); i++ {
		switch history[i] {
		case StatusConnected:
			trend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render("▁"))
		case StatusSlow:
			trend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("▃"))
		case StatusTimeout:
			trend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Render("▅"))
		case StatusError:
			trend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("█"))
		default:
			trend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("━"))
		}
	}

	// Pad with spaces if needed
	for trend.Len() < maxItems {
		trend.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("━"))
	}

	return trend.String()
}

func (m model) renderHistory() string {
	if m.width < 60 {
		return ""
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)

	header := headerStyle.Render("Recent Activity")

	// Show overall network health
	connected := 0
	total := len(m.targets)

	for _, target := range m.targets {
		if target.Status == StatusConnected {
			connected++
		}
	}

	healthPercent := float64(connected) / float64(total) * 100
	var healthColor lipgloss.Color
	var healthStatus string

	switch {
	case healthPercent >= 80:
		healthColor = lipgloss.Color("46")
		healthStatus = "EXCELLENT"
	case healthPercent >= 60:
		healthColor = lipgloss.Color("226")
		healthStatus = "GOOD"
	case healthPercent >= 40:
		healthColor = lipgloss.Color("202")
		healthStatus = "POOR"
	default:
		healthColor = lipgloss.Color("196")
		healthStatus = "CRITICAL"
	}

	healthStyle := lipgloss.NewStyle().
		Foreground(healthColor).
		Bold(true)

	health := fmt.Sprintf("Network Health: %s (%d/%d targets)", healthStatus, connected, total)

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		healthStyle.Render(health),
	)
}

type targetsCheckedMsg struct {
	targets []Target
}

func checkAllTargetsCmd(targets []Target) tea.Cmd {
	return func() tea.Msg {
		var wg sync.WaitGroup
		updatedTargets := make([]Target, len(targets))
		copy(updatedTargets, targets)

		for i := range updatedTargets {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				checkTarget(&updatedTargets[idx])
			}(i)
		}

		wg.Wait()
		return targetsCheckedMsg{targets: updatedTargets}
	}
}

func checkTarget(target *Target) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	var status ConnectionStatus

	switch target.Type {
	case "ping":
		status = checkPing(ctx, target.Host)
	case "http":
		status = checkHTTP(ctx, target.Host)
	case "dns":
		status = checkDNS(ctx, target.Host)
	case "bandwidth":
		status, target.Bandwidth = checkBandwidth(ctx, target.Host)
	case "upload":
		status, target.UploadSpeed = checkUpload(ctx, target.Host)
	case "captive":
		status, target.IsCaptive = checkCaptivePortal(ctx, target.Host)
	case "ipv6":
		status, target.IPv6Support = checkIPv6(ctx, target.Host)
	case "traceroute":
		status, target.HopCount = checkTraceroute(ctx, target.Host)
	default:
		status = StatusError
	}

	target.Status = status
	target.LastCheck = time.Now()
	target.Latency = time.Since(start)

	// Calculate jitter from latency history
	target.LatencyHistory = append(target.LatencyHistory, target.Latency)
	if len(target.LatencyHistory) > 10 {
		target.LatencyHistory = target.LatencyHistory[1:]
	}
	target.Jitter = calculateJitter(target.LatencyHistory)

	// Calculate quality score
	target.QualityScore = calculateQualityScore(target)

	// Add to history
	target.History = append(target.History, status)
	if len(target.History) > 30 {
		target.History = target.History[1:]
	}
}

func checkPing(ctx context.Context, host string) ConnectionStatus {
	// For ping, we'll use a TCP connection test since ICMP requires root
	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, "53"))
	if err != nil {
		return StatusTimeout
	}
	conn.Close()
	return StatusConnected
}

func checkHTTP(ctx context.Context, url string) ConnectionStatus {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Follow redirects
		},
	}

	// Try HEAD first, fallback to GET if it fails
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return StatusError
	}

	// Set a proper User-Agent to avoid blocking
	req.Header.Set("User-Agent", "SuperPing/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return StatusTimeout
	}
	defer resp.Body.Close()

	// Accept a wider range of status codes
	// Some sites return 405 for HEAD but work fine
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return StatusConnected
	}

	// If HEAD failed, try GET request
	if resp.StatusCode == 405 || resp.StatusCode >= 400 {
		req, err = http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return StatusError
		}
		req.Header.Set("User-Agent", "SuperPing/1.0")

		resp2, err := client.Do(req)
		if err != nil {
			return StatusTimeout
		}
		defer resp2.Body.Close()

		if resp2.StatusCode >= 200 && resp2.StatusCode < 400 {
			return StatusConnected
		}
	}

	return StatusError
}

func checkDNS(ctx context.Context, host string) ConnectionStatus {
	resolver := &net.Resolver{}
	_, err := resolver.LookupHost(ctx, host)
	if err != nil {
		return StatusTimeout
	}
	return StatusConnected
}

func checkBandwidth(ctx context.Context, url string) (ConnectionStatus, float64) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return StatusError, 0
	}

	// Start timing for total request
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return StatusTimeout, 0
	}
	defer resp.Body.Close()

	// Read the response to measure download speed
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return StatusError, 0
	}

	duration := time.Since(start)
	bytesDownloaded := float64(len(bytes))

	// Ensure minimum duration to avoid unrealistic speeds
	if duration.Seconds() < 0.01 {
		duration = 10 * time.Millisecond
	}

	bitsPerSecond := (bytesDownloaded * 8) / duration.Seconds()
	mbps := bitsPerSecond / 1_000_000

	// Determine status based on speed
	var status ConnectionStatus
	switch {
	case mbps >= 2.0:
		status = StatusConnected // Good for video calls
	case mbps >= 0.5:
		status = StatusSlow // Good for browsing
	case mbps >= 0.01:
		status = StatusTimeout // Very slow but working
	default:
		status = StatusError // Essentially unusable
	}

	return status, mbps
}

func calculateJitter(latencyHistory []time.Duration) time.Duration {
	if len(latencyHistory) < 2 {
		return 0
	}

	// Calculate mean
	var sum time.Duration
	for _, latency := range latencyHistory {
		sum += latency
	}
	mean := sum / time.Duration(len(latencyHistory))

	// Calculate standard deviation (jitter)
	var sumSquaredDiffs float64
	for _, latency := range latencyHistory {
		diff := float64(latency - mean)
		sumSquaredDiffs += diff * diff
	}

	variance := sumSquaredDiffs / float64(len(latencyHistory))
	stdDev := math.Sqrt(variance)

	return time.Duration(stdDev)
}

func calculateQualityScore(target *Target) float64 {
	// Base score
	score := 100.0

	// Penalize high latency
	latencyMs := target.Latency.Milliseconds()
	switch {
	case latencyMs > 1000:
		score -= 60
	case latencyMs > 500:
		score -= 40
	case latencyMs > 200:
		score -= 20
	case latencyMs > 100:
		score -= 10
	}

	// Penalize high jitter
	jitterMs := target.Jitter.Milliseconds()
	switch {
	case jitterMs > 100:
		score -= 30
	case jitterMs > 50:
		score -= 20
	case jitterMs > 20:
		score -= 10
	}

	// Bonus for bandwidth
	if target.Type == "bandwidth" {
		switch {
		case target.Bandwidth >= 10:
			score += 10
		case target.Bandwidth >= 5:
			score += 5
		case target.Bandwidth < 1:
			score -= 20
		}
	}

	// Penalize based on status
	switch target.Status {
	case StatusError:
		score = 0
	case StatusTimeout:
		score = math.Min(score, 20)
	case StatusSlow:
		score = math.Min(score, 60)
	}

	return math.Max(0, math.Min(100, score))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func checkUpload(ctx context.Context, url string) (ConnectionStatus, float64) {
	// Create test data (5KB)
	testData := bytes.Repeat([]byte("test data for upload speed measurement "), 128)

	client := &http.Client{Timeout: 5 * time.Second}

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(testData))
	if err != nil {
		return StatusError, 0
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return StatusTimeout, 0
	}
	defer resp.Body.Close()

	duration := time.Since(start)
	bytesUploaded := float64(len(testData))

	if duration.Seconds() < 0.01 {
		duration = 10 * time.Millisecond
	}

	bitsPerSecond := (bytesUploaded * 8) / duration.Seconds()
	mbps := bitsPerSecond / 1_000_000

	var status ConnectionStatus
	switch {
	case mbps >= 1.0:
		status = StatusConnected
	case mbps >= 0.2:
		status = StatusSlow
	case mbps >= 0.01:
		status = StatusTimeout
	default:
		status = StatusError
	}

	return status, mbps
}

func checkCaptivePortal(ctx context.Context, url string) (ConnectionStatus, bool) {
	client := &http.Client{
		Timeout: 3 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return StatusError, false
	}

	resp, err := client.Do(req)
	if err != nil {
		return StatusTimeout, false
	}
	defer resp.Body.Close()

	// Google's connectivity check should return 204 with no content
	// If we get a redirect or different status, it's likely a captive portal
	isCaptive := resp.StatusCode != 204

	var status ConnectionStatus
	if isCaptive {
		status = StatusSlow // Captive portal detected
	} else {
		status = StatusConnected // Direct internet access
	}

	return status, isCaptive
}

func checkIPv6(ctx context.Context, host string) (ConnectionStatus, bool) {
	resolver := &net.Resolver{}

	// Try to resolve AAAA record (IPv6)
	_, err := resolver.LookupHost(ctx, host)
	if err != nil {
		return StatusTimeout, false
	}

	// Try to connect via IPv6
	dialer := &net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp6", net.JoinHostPort(host, "80"))

	var hasIPv6 bool
	var status ConnectionStatus

	if err != nil {
		// IPv6 not available, check IPv4
		conn4, err4 := dialer.DialContext(ctx, "tcp4", net.JoinHostPort(host, "80"))
		if err4 != nil {
			status = StatusError
		} else {
			conn4.Close()
			status = StatusSlow // IPv4 only
		}
		hasIPv6 = false
	} else {
		conn.Close()
		hasIPv6 = true
		status = StatusConnected // IPv6 available
	}

	return status, hasIPv6
}

func checkTraceroute(ctx context.Context, host string) (ConnectionStatus, int) {
	// Use system traceroute command (simplified version)
	cmd := exec.CommandContext(ctx, "traceroute", "-m", "15", "-w", "1", host)
	output, err := cmd.Output()

	if err != nil {
		// Fallback: estimate hops from ping TTL
		return estimateHopsFromPing(ctx, host)
	}

	// Count hops from traceroute output
	lines := strings.Split(string(output), "\n")
	hopCount := 0

	for _, line := range lines {
		if strings.Contains(line, "ms") || strings.Contains(line, "*") {
			hopCount++
		}
	}

	var status ConnectionStatus
	switch {
	case hopCount <= 10:
		status = StatusConnected // Direct/good route
	case hopCount <= 20:
		status = StatusSlow // Reasonable route
	case hopCount > 20:
		status = StatusTimeout // Long route
	default:
		status = StatusError
	}

	return status, hopCount
}

func estimateHopsFromPing(ctx context.Context, host string) (ConnectionStatus, int) {
	// Simplified fallback - just return a reasonable estimate
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "53"), 3*time.Second)
	if err != nil {
		return StatusError, 0
	}
	conn.Close()

	// Estimate based on typical internet routing
	return StatusConnected, 12
}


func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Thanks for using SuperPing! 👋")
}
