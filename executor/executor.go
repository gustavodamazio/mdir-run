package executor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gustavodamazio/mdir-run/config"
	"github.com/gustavodamazio/mdir-run/logger"
	"github.com/gustavodamazio/mdir-run/progress"
)

func executeWithRetry(cmd *exec.Cmd, stdoutBuf, stderrBuf *bytes.Buffer, retries int) (int, error) {
	var err error
	for attempt := 0; attempt <= retries; attempt++ {
		// Reset buffers before each attempt
		stdoutBuf.Reset()
		stderrBuf.Reset()

		err = cmd.Run()
		if err == nil {
			return attempt + 1, nil // Command succeeded, return attempt number (1-indexed)
		}

		// Don't sleep after the last attempt
		if attempt < retries {
			time.Sleep(time.Second * time.Duration(attempt+1)) // Simple linear backoff
		}
	}
	return retries + 1, err // Return the last attempt number and last error
}

func ProcessRepo(dir string, cfg *config.Config, progressManager *progress.ProgressManager) {
	progress := progressManager.GetProgress(dir)
	startTime := time.Now()
	status := "SUCCESS"

	dirPath := filepath.Join(cfg.InitialDir, dir)
	stat, err := os.Stat(dirPath)
	if err != nil || !stat.IsDir() {
		progress.Status = "FAIL"
		errorMsg := fmt.Sprintf("Failed to access directory: %s", dirPath)
		progress.Command = errorMsg

		// Write detailed error log for directory access failure
		var errorDetails string
		if err != nil {
			errorDetails = fmt.Sprintf("Error: %v", err)
		} else {
			errorDetails = "Path exists but is not a directory"
		}
		logger.WriteErrorLog(cfg.LogFile, dir, errorDetails)

		logger.WriteLog(cfg.LogFile, progress.Status, time.Since(startTime).Seconds(), dir)

		progressManager.UpdateProgress(dir, progress)
		return
	}

	// Check if 'SubDirsEntryPoints' directories exist
	for _, subDir := range cfg.SubDirsEntryPoints {
		subDirPath := filepath.Join(dirPath, subDir)
		if stat, err := os.Stat(subDirPath); err == nil && stat.IsDir() {
			dirPath = subDirPath
			break
		}
	}

	progress.Total = len(cfg.Commands)
	var successDetail strings.Builder
	successDetail.WriteString(fmt.Sprintf("Working directory: %s\n\n", dirPath))

	for i, cmdArgs := range cfg.Commands {
		progress.Step = i + 1
		cmdString := strings.Join(cmdArgs, " ")
		progress.Command = cmdString
		progressManager.UpdateProgress(dir, progress)

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = dirPath

		// Capture the output
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		attemptNumber, err := executeWithRetry(cmd, &stdoutBuf, &stderrBuf, cfg.Retries)
		if err != nil {
			progress.Status = fmt.Sprintf("FAIL(%d/%d)", attemptNumber, cfg.Retries+1)
			errorOutput := stderrBuf.String()
			progress.Output = errorOutput
			progress.Command = fmt.Sprintf("Failed to execute %s", progress.Command)

			// Write detailed error log to separate file
			// Include both stdout and stderr in the error log
			errorDetails := fmt.Sprintf("Command: %s\nError: %v\nAttempt: %d/%d\nStderr Output:\n%s\nStdout Output:\n%s",
				progress.Command, err, attemptNumber, cfg.Retries+1, errorOutput, stdoutBuf.String())
			logger.WriteErrorLog(cfg.LogFile, dir, errorDetails)

			progressManager.UpdateProgress(dir, progress)
			break
		} 
		// Always show the attempt count for SUCCESS, regardless of retry count
		status = fmt.Sprintf("SUCCESS(%d/%d)", attemptNumber, cfg.Retries+1)

		// Add command execution details for success log
		successDetail.WriteString(fmt.Sprintf("Command %d/%d: %s\n", i+1, progress.Total, cmdString))
		successDetail.WriteString(fmt.Sprintf("Attempts needed: %d/%d\n", attemptNumber, cfg.Retries+1))

		if stdoutBuf.Len() > 0 {
			successDetail.WriteString(fmt.Sprintf("Stdout Output:\n%s\n", stdoutBuf.String()))
		} else {
			successDetail.WriteString("Stdout: No output\n")
		}

		if stderrBuf.Len() > 0 {
			successDetail.WriteString(fmt.Sprintf("Stderr Output:\n%s\n", stderrBuf.String()))
		}

		successDetail.WriteString("\n---\n\n")
	}

	executionTime := time.Since(startTime).Seconds()
	if progress.Status != "FAIL" && !strings.HasPrefix(progress.Status, "FAIL(") {
		progress.Status = status

		// Write success log with execution details
		successDetail.WriteString(fmt.Sprintf("\nExecution completed in %.2f seconds", executionTime))
		logger.WriteSuccessLog(cfg.LogFile, dir, successDetail.String())
	}

	logger.WriteLog(cfg.LogFile, progress.Status, executionTime, dir)
	progressManager.UpdateProgress(dir, progress)
}
