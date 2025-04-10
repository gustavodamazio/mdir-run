package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var logMutex sync.Mutex

func InitializeLogFile(logFile string) error {
	// Ensure parent directory exists
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

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

func WriteErrorLog(logFile, dir string, errorDetails string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	logDir := filepath.Dir(logFile)
	errorFileName := filepath.Join(logDir, fmt.Sprintf("%s_error.txt", filepath.Base(dir)))
	
	f, err := os.Create(errorFileName)
	if err != nil {
		fmt.Printf("%s | ERROR: Failed to create error log file: %s\n", dir, err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("02/01/2006 15:04:05")
	header := fmt.Sprintf("Error log for directory '%s' created on %s\n\n", dir, timestamp)
	
	// Add debugging info when content might be empty
	if strings.TrimSpace(errorDetails) == "" || !strings.Contains(errorDetails, "Output") {
		errorDetails += "\n\nNOTE: Command output was empty. This might indicate:\n" +
			"- Command did not produce any output before failing\n" +
			"- Command may have written output to a file instead of stdout/stderr\n" +
			"- There might be an environment or permission issue\n" +
			"- The process might have been terminated by the operating system"
	}
	
	content := header + errorDetails
	
	if _, err := f.WriteString(content); err != nil {
		fmt.Printf("%s | ERROR: Failed to write to error log file: %s\n", dir, err)
	}
}

func WriteSuccessLog(logFile, dir string, executionDetails string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	logDir := filepath.Dir(logFile)
	successFileName := filepath.Join(logDir, fmt.Sprintf("%s_success.txt", filepath.Base(dir)))
	
	f, err := os.Create(successFileName)
	if err != nil {
		fmt.Printf("%s | ERROR: Failed to create success log file: %s\n", dir, err)
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("02/01/2006 15:04:05")
	header := fmt.Sprintf("Success log for directory '%s' created on %s\n\n", dir, timestamp)
	
	content := header + executionDetails
	
	if _, err := f.WriteString(content); err != nil {
		fmt.Printf("%s | ERROR: Failed to write to success log file: %s\n", dir, err)
	}
}

// WriteSummaryLog writes a final summary line to the log file with the execution end date and total time
func WriteSummaryLog(logFile string, startTime time.Time) {
	logMutex.Lock()
	defer logMutex.Unlock()

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("ERROR: Failed to open log file for summary: %s\n", err)
		return
	}
	defer f.Close()

	endTime := time.Now()
	executionDuration := endTime.Sub(startTime)
	
	// Format duration as minutes and seconds (e.g., "3m 5s")
	minutes := int(executionDuration.Minutes())
	seconds := int(executionDuration.Seconds()) % 60
	durationStr := fmt.Sprintf("%dm %ds", minutes, seconds)
	
	summaryLine := fmt.Sprintf("\nExecution completed on %s | Total execution time: %s\n", 
		endTime.Format("02/01/2006 15:04:05"), durationStr)
	
	if _, err := f.WriteString(summaryLine); err != nil {
		fmt.Printf("ERROR: Failed to write summary to log file: %s\n", err)
	}
}

