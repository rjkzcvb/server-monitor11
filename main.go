package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server-monitor/monitor"
)

func main() {
	// Создаем монитор
	monitor := monitor.NewMonitor("http://srv.msk01.gigacorp.local/_stats")
	
	// Канал для обработки сигналов завершения
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)
	
	// Тикер для периодического опроса (например, каждые 10 секунд)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	fmt.Println("Starting server monitor...")
	fmt.Println("Press Ctrl+C to stop")
	
	for {
		select {
		case <-ticker.C:
			monitor.CheckStats()
		case <-interrupt:
			fmt.Println("\nShutting down monitor...")
			return
		}
	}
}
