package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type ServerConfig struct {
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // "tcp", "http", "https"
	Timeout  int    `json:"timeout"`  // seconds
}

type HealthResult struct {
	Server      ServerConfig `json:"server"`
	Status      string       `json:"status"`      // "UP", "DOWN"
	ResponseTime int64       `json:"response_time"` // milliseconds
	Timestamp   time.Time    `json:"timestamp"`
	Error       string       `json:"error,omitempty"`
}

type Monitor struct {
	servers []ServerConfig
	results chan HealthResult
	wg      sync.WaitGroup
}

func NewMonitor() *Monitor {
	return &Monitor{
		results: make(chan HealthResult, 100),
	}
}

func (m *Monitor) LoadConfig(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var config struct {
		Servers []ServerConfig `json:"servers"`
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	m.servers = config.Servers
	return nil
}

func (m *Monitor) checkTCP(server ServerConfig) HealthResult {
	start := time.Now()
	address := net.JoinHostPort(server.Host, strconv.Itoa(server.Port))
	
	conn, err := net.DialTimeout("tcp", address, time.Duration(server.Timeout)*time.Second)
	responseTime := time.Since(start).Milliseconds()
	
	result := HealthResult{
		Server:       server,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
	}

	if err != nil {
		result.Status = "DOWN"
		result.Error = err.Error()
	} else {
		result.Status = "UP"
		conn.Close()
	}

	return result
}

func (m *Monitor) checkHTTP(server ServerConfig) HealthResult {
	start := time.Now()
	url := fmt.Sprintf("%s://%s:%d", server.Protocol, server.Host, server.Port)
	
	client := &http.Client{
		Timeout: time.Duration(server.Timeout) * time.Second,
	}

	resp, err := client.Get(url)
	responseTime := time.Since(start).Milliseconds()
	
	result := HealthResult{
		Server:       server,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
	}

	if err != nil {
		result.Status = "DOWN"
		result.Error = err.Error()
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			result.Status = "UP"
		} else {
			result.Status = "DOWN"
			result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	return result
}

func (m *Monitor) checkServer(server ServerConfig) {
	defer m.wg.Done()
	
	var result HealthResult
	
	switch server.Protocol {
	case "tcp":
		result = m.checkTCP(server)
	case "http", "https":
		result = m.checkHTTP(server)
	default:
		result = HealthResult{
			Server:    server,
			Status:    "DOWN",
			Timestamp: time.Now(),
			Error:     "unsupported protocol: " + server.Protocol,
		}
	}
	
	m.results <- result
}

func (m *Monitor) RunCheck() {
	fmt.Printf("Checking %d servers...\n", len(m.servers))
	
	// Start goroutines for concurrent checking
	for _, server := range m.servers {
		m.wg.Add(1)
		go m.checkServer(server)
	}

	// Close results channel when all checks complete
	go func() {
		m.wg.Wait()
		close(m.results)
	}()

	// Collect and display results
	var upCount, downCount int
	for result := range m.results {
		status := "✓"
		if result.Status == "DOWN" {
			status = "✗"
			downCount++
		} else {
			upCount++
		}

		fmt.Printf("%s [%s] %s:%d - %s (%dms)",
			status, result.Status, result.Server.Host, result.Server.Port,
			result.Server.Name, result.ResponseTime)
		
		if result.Error != "" {
			fmt.Printf(" - Error: %s", result.Error)
		}
		fmt.Println()
	}

	fmt.Printf("\nSummary: %d UP, %d DOWN\n", upCount, downCount)
}

func (m *Monitor) StartContinuousMonitoring(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	fmt.Printf("Starting continuous monitoring (interval: %v)\n", interval)
	fmt.Println("Press Ctrl+C to stop...")

	for {
		select {
		case <-ticker.C:
			fmt.Printf("\n--- Health Check at %s ---\n", time.Now().Format("15:04:05"))
			m.RunCheck()
		}
	}
}

func (m *Monitor) GenerateReport(filename string) error {
	// Run a single check
	m.RunCheck()
	
	// Collect results for report
	var results []HealthResult
	for _, server := range m.servers {
		m.wg.Add(1)
		go m.checkServer(server)
	}

	go func() {
		m.wg.Wait()
		close(m.results)
	}()

	for result := range m.results {
		results = append(results, result)
	}

	// Generate JSON report
	report := struct {
		Timestamp time.Time      `json:"timestamp"`
		Results   []HealthResult `json:"results"`
		Summary   struct {
			Total int `json:"total"`
			Up    int `json:"up"`
			Down  int `json:"down"`
		} `json:"summary"`
	}{
		Timestamp: time.Now(),
		Results:   results,
	}

	for _, result := range results {
		report.Summary.Total++
		if result.Status == "UP" {
			report.Summary.Up++
		} else {
			report.Summary.Down++
		}
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func createSampleConfig() {
	config := struct {
		Servers []ServerConfig `json:"servers"`
	}{
		Servers: []ServerConfig{
			{Name: "Google DNS", Host: "8.8.8.8", Port: 53, Protocol: "tcp", Timeout: 5},
			{Name: "Google", Host: "google.com", Port: 80, Protocol: "http", Timeout: 10},
			{Name: "GitHub", Host: "github.com", Port: 443, Protocol: "https", Timeout: 10},
			{Name: "Local SSH", Host: "localhost", Port: 22, Protocol: "tcp", Timeout: 3},
			{Name: "Local Web", Host: "localhost", Port: 8080, Protocol: "http", Timeout: 5},
		},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile("servers.json", data, 0644)
	fmt.Println("Created sample configuration: servers.json")
}

func printUsage() {
	fmt.Println("Server Health Monitor")
	fmt.Println("Usage:")
	fmt.Println("  go run main.go [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -config <file>     Configuration file (default: servers.json)")
	fmt.Println("  -once             Run check once and exit")
	fmt.Println("  -interval <dur>   Continuous monitoring interval (default: 30s)")
	fmt.Println("  -report <file>    Generate JSON report")
	fmt.Println("  -sample           Create sample configuration file")
	fmt.Println("  -help             Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  go run main.go -sample")
	fmt.Println("  go run main.go -once")
	fmt.Println("  go run main.go -interval 60s")
	fmt.Println("  go run main.go -report health_report.json")
}

func main() {
	args := os.Args[1:]
	
	configFile := "servers.json"
	runOnce := false
	interval := 30 * time.Second
	reportFile := ""

	// Simple argument parsing
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-help", "--help", "-h":
			printUsage()
			return
		case "-sample":
			createSampleConfig()
			return
		case "-config":
			if i+1 < len(args) {
				configFile = args[i+1]
				i++
			}
		case "-once":
			runOnce = true
		case "-interval":
			if i+1 < len(args) {
				if d, err := time.ParseDuration(args[i+1]); err == nil {
					interval = d
				}
				i++
			}
		case "-report":
			if i+1 < len(args) {
				reportFile = args[i+1]
				i++
			}
		}
	}

	monitor := NewMonitor()

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Printf("Config file '%s' not found. Creating sample...\n", configFile)
		createSampleConfig()
	}

	if err := monitor.LoadConfig(configFile); err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("Loaded %d servers from %s\n", len(monitor.servers), configFile)
	fmt.Printf("Go version: %s, OS: %s, Arch: %s\n", 
		runtime.Version(), runtime.GOOS, runtime.GOARCH)

	if reportFile != "" {
		fmt.Printf("Generating report: %s\n", reportFile)
		if err := monitor.GenerateReport(reportFile); err != nil {
			log.Fatalf("Error generating report: %v", err)
		}
		fmt.Printf("Report saved to %s\n", reportFile)
	} else if runOnce {
		monitor.RunCheck()
	} else {
		monitor.StartContinuousMonitoring(interval)
	}
}