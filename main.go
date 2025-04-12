package main

import (
	"flag"
	"log"
	"sync"
	"time"

	"github.com/gustavodamazio/mdir-run/config"
	"github.com/gustavodamazio/mdir-run/directories"
	"github.com/gustavodamazio/mdir-run/executor"
	"github.com/gustavodamazio/mdir-run/gui"
	"github.com/gustavodamazio/mdir-run/logger"
	"github.com/gustavodamazio/mdir-run/progress"

	"github.com/gosuri/uilive"
)

func main() {
	// Add a flag to enable GUI mode 
	guiFlag := flag.Bool("gui", true, "Enable GUI mode (default: true, use -gui=false for CLI mode)")
	cliFlag := flag.Bool("cli", false, "Force CLI mode instead of GUI mode")
	
	// Now parse all flags
	flag.Parse()
	
	// Check if any CLI-specific flags were provided
	hasCLIFlags := *cliFlag || 
		*config.DirFlag != "" || 
		*config.CommandsFlag != "" || 
		*config.SubDirsFlag != "" || 
		*config.ConcurrencyFlag != 10 ||  // 10 is the default
		*config.RetriesFlag != 0       // 0 is the default
	
	// If CLI flags were provided or CLI mode is explicitly requested, use CLI mode
	if hasCLIFlags || !(*guiFlag) {
		runCLIMode()
		return
	}
	
	// If no CLI flags were provided and GUI is not disabled, launch the GUI
	gui.LaunchGUI()
}

func runCLIMode() {
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
	if _, err := logger.ArchiveLogs(cfg.LogFile); err != nil {
		log.Printf("WARNING: Failed to archive log files: %v", err)
	}
}
