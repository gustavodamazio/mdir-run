package main

import (
	"log"
	"sync"
	"time"

	"github.com/gustavodamazio/mdir-run/config"
	"github.com/gustavodamazio/mdir-run/directories"
	"github.com/gustavodamazio/mdir-run/executor"
	"github.com/gustavodamazio/mdir-run/logger"
	"github.com/gustavodamazio/mdir-run/progress"

	"github.com/gosuri/uilive"
)

func main() {
	// Record the start time for the overall execution
	startTime := time.Now()
	
	// Parse configuration
	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Initialize the log file
	err = logger.InitializeLogFile(cfg.LogFile)
	if err != nil {
		log.Fatalf("Failed to initialize log file: %v", err)
	}

	// Get list of directories to process
	dirs, err := directories.GetDirectories(cfg.InitialDir)
	if err != nil {
		log.Fatalf("Failed to get directories: %v", err)
	}

	// Initialize progress manager
	progressManager := progress.NewProgressManager(dirs)

	// Initialize the writer
	writer := uilive.New()
	writer.Start()
	defer writer.Stop()

	// Limit concurrency
	semaphore := make(chan struct{}, cfg.Concurrency)
	var wg sync.WaitGroup

	// Process directories
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			executor.ProcessRepo(dir, cfg, progressManager)
		}(dir)
	}

	// Start display updater
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				progressManager.PrintAllProgress(writer)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	close(done)
	progressManager.PrintAllProgress(writer)
	
	// Write the final summary to the log file with execution time
	logger.WriteSummaryLog(cfg.LogFile, startTime)
	
	// Archive log files and remove originals
	if err := logger.ArchiveLogs(cfg.LogFile); err != nil {
		log.Printf("WARNING: Failed to archive log files: %v", err)
	}
}
