package logger

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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

// ArchiveLogs archives all log files into a compressed archive and removes the original files
// Returns the archive path and any error
func ArchiveLogs(logFile string) (string, error) {
	logMutex.Lock()
	defer logMutex.Unlock()

	// Get the directory of the main log file
	logDir := filepath.Dir(logFile)
	
	// Generate timestamp for archive name
	timestamp := time.Now().Format("20060102-150405")
	
	// Define archive filename based on OS
	var archivePath string
	isWindows := runtime.GOOS == "windows"
	
	if isWindows {
		archivePath = filepath.Join(logDir, fmt.Sprintf("logs-%s.zip", timestamp))
	} else {
		archivePath = filepath.Join(logDir, fmt.Sprintf("logs-%s.tar.gz", timestamp))
	}
	
	// Collect all log files to be archived
	logFiles := []string{}
	
	// Add main log file
	if _, err := os.Stat(logFile); err == nil {
		logFiles = append(logFiles, logFile)
	}
	
	// Find all success and error log files in the same directory
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return "", fmt.Errorf("failed to read log directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		fileName := entry.Name()
		if strings.HasSuffix(fileName, "_success.txt") || strings.HasSuffix(fileName, "_error.txt") {
			logFiles = append(logFiles, filepath.Join(logDir, fileName))
		}
	}
	
	// If no log files found, return
	if len(logFiles) == 0 {
		return "", fmt.Errorf("no log files found to archive")
	}
	
	// Create archive based on OS
	var archiveErr error
	if isWindows {
		archiveErr = createZipArchive(archivePath, logFiles)
	} else {
		archiveErr = createTarGzArchive(archivePath, logFiles)
	}
	
	if archiveErr != nil {
		return "", fmt.Errorf("failed to create archive: %w", archiveErr)
	}
	
	// Delete the original log files
	for _, file := range logFiles {
		if err := os.Remove(file); err != nil {
			fmt.Printf("WARNING: Failed to delete original log file %s: %s\n", file, err)
		}
	}
	
	fmt.Printf("Log files archived to %s\n", archivePath)
	return archivePath, nil
}

// createZipArchive creates a zip archive containing the specified files
func createZipArchive(archivePath string, files []string) error {
	zipFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer zipFile.Close()
	
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	for _, file := range files {
		if err := addFileToZip(zipWriter, file); err != nil {
			return fmt.Errorf("failed to add file to zip: %w", err)
		}
	}
	
	return nil
}

// addFileToZip adds a file to a zip archive
func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return fmt.Errorf("failed to create file header: %w", err)
	}
	
	// Use just the base filename to avoid directory structure in the archive
	header.Name = filepath.Base(filePath)
	
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create file in zip: %w", err)
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to write file to zip: %w", err)
	}
	
	return nil
}

// createTarGzArchive creates a tar.gz archive containing the specified files
func createTarGzArchive(archivePath string, files []string) error {
	tarFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %w", err)
	}
	defer tarFile.Close()
	
	gzWriter := gzip.NewWriter(tarFile)
	defer gzWriter.Close()
	
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()
	
	for _, file := range files {
		if err := addFileToTar(tarWriter, file); err != nil {
			return fmt.Errorf("failed to add file to tar: %w", err)
		}
	}
	
	return nil
}

// addFileToTar adds a file to a tar archive
func addFileToTar(tarWriter *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	
	header, err := tar.FileInfoHeader(fileInfo, "")
	if err != nil {
		return fmt.Errorf("failed to create tar header: %w", err)
	}
	
	// Use just the base filename to avoid directory structure in the archive
	header.Name = filepath.Base(filePath)
	
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}
	
	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf("failed to write file content to tar: %w", err)
	}
	
	return nil
}

