package monitor

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ServerStats представляет статистику сервера
type ServerStats struct {
	LoadAverage        float64
	TotalMemory        uint64
	UsedMemory         uint64
	TotalDisk          uint64
	UsedDisk           uint64
	NetworkBandwidth   uint64
	NetworkUsage       uint64
}

// Monitor представляет монитор сервера
type Monitor struct {
	serverURL string
	errorCount int
	client    *http.Client
}

// NewMonitor создает новый монитор
func NewMonitor(serverURL string) *Monitor {
	return &Monitor{
		serverURL: serverURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckStats выполняет проверку статистики сервера
func (m *Monitor) CheckStats() {
	stats, err := m.fetchStats()
	if err != nil {
		m.handleError(err)
		return
	}
	
	// Сбрасываем счетчик ошибок при успешном запросе
	m.errorCount = 0
	
	// Проверяем пороги
	m.checkThresholds(stats)
}

// fetchStats получает статистику с сервера
func (m *Monitor) fetchStats() (*ServerStats, error) {
	resp, err := m.client.Get(m.serverURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stats: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status: %s", resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	return m.parseStats(string(body))
}

// parseStats парсит статистику из строки
func (m *Monitor) parseStats(data string) (*ServerStats, error) {
	parts := strings.Split(strings.TrimSpace(data), ",")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid data format: expected 7 values, got %d", len(parts))
	}
	
	var stats ServerStats
	var err error
	
	// Load Average
	stats.LoadAverage, err = strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, fmt.Errorf("invalid load average: %v", err)
	}
	
	// Total Memory
	stats.TotalMemory, err = strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid total memory: %v", err)
	}
	
	// Used Memory
	stats.UsedMemory, err = strconv.ParseUint(parts[2], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid used memory: %v", err)
	}
	
	// Total Disk
	stats.TotalDisk, err = strconv.ParseUint(parts[3], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid total disk: %v", err)
	}
	
	// Used Disk
	stats.UsedDisk, err = strconv.ParseUint(parts[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid used disk: %v", err)
	}
	
	// Network Bandwidth
	stats.NetworkBandwidth, err = strconv.ParseUint(parts[5], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid network bandwidth: %v", err)
	}
	
	// Network Usage
	stats.NetworkUsage, err = strconv.ParseUint(parts[6], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid network usage: %v", err)
	}
	
	return &stats, nil
}

// checkThresholds проверяет пороговые значения
func (m *Monitor) checkThresholds(stats *ServerStats) {
	// Проверка Load Average
	if stats.LoadAverage > 30 {
		fmt.Printf("Load Average is too high: %.2f\n", stats.LoadAverage)
	}
	
	// Проверка использования памяти
	if stats.TotalMemory > 0 {
		memoryUsage := float64(stats.UsedMemory) / float64(stats.TotalMemory) * 100
		if memoryUsage > 80 {
			fmt.Printf("Memory usage too high: %.2f%%\n", memoryUsage)
		}
	}
	
	// Проверка дискового пространства
	if stats.TotalDisk > 0 {
		diskUsage := float64(stats.UsedDisk) / float64(stats.TotalDisk) * 100
		if diskUsage > 90 {
			freeSpaceMB := float64(stats.TotalDisk-stats.UsedDisk) / 1024 / 1024
			fmt.Printf("Free disk space is too low: %.2f Mb left\n", freeSpaceMB)
		}
	}
	
	// Проверка использования сети
	if stats.NetworkBandwidth > 0 {
		networkUsage := float64(stats.NetworkUsage) / float64(stats.NetworkBandwidth) * 100
		if networkUsage > 90 {
			availableBandwidthMbit := float64(stats.NetworkBandwidth-stats.NetworkUsage) * 8 / 1000000
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", availableBandwidthMbit)
		}
	}
}

// handleError обрабатывает ошибки
func (m *Monitor) handleError(err error) {
	m.errorCount++
	
	if m.errorCount >= 3 {
		fmt.Println("Unable to fetch server statistic")
	}
}
