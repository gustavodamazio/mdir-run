package logger

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var logMutex sync.Mutex

func InitializeLogFile(logFile string) error {
	f, err := os.Create(logFile)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer f.Close()

	// Write header to log file
	header := fmt.Sprintf("Script execution log on %s\n", time.Now().Format("02/01/2006 15:04:05"))
	if _, err := f.WriteString(header); err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}
	return nil
}

func WriteLog(logFile, status string, executionTime float64, dir string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("%s | ERROR: Failed to open log file: %s\n", dir, err)
		return
	}
	defer f.Close()

	logLine := fmt.Sprintf("STATUS: %-10s | TIME: %5.0f sec | DIR: %-30s\n", status, executionTime, dir)
	if _, err := f.WriteString(logLine); err != nil {
		fmt.Printf("%s | ERROR: Failed to write to log file: %s\n", dir, err)
	}
}
