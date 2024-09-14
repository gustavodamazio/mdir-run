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

func ProcessRepo(dir string, cfg *config.Config, progressManager *progress.ProgressManager) {
	progress := progressManager.GetProgress(dir)
	startTime := time.Now()
	status := "SUCCESS"

	dirPath := filepath.Join(cfg.InitialDir, dir)
	stat, err := os.Stat(dirPath)
	if err != nil || !stat.IsDir() {
		progress.Status = "FAIL"
		progress.Command = fmt.Sprintf("Failed to access directory: %s", dirPath)
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
	for i, cmdArgs := range cfg.Commands {
		progress.Step = i + 1
		progress.Command = strings.Join(cmdArgs, " ")
		progressManager.UpdateProgress(dir, progress)

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = dirPath

		// Capture the output
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = &stdoutBuf
		cmd.Stderr = &stderrBuf

		if err := cmd.Run(); err != nil {
			progress.Status = "FAIL"
			progress.Output = stderrBuf.String()
			progress.Command = fmt.Sprintf("Failed to execute %s", progress.Command)
			progressManager.UpdateProgress(dir, progress)
			break
		}
	}

	executionTime := time.Since(startTime).Seconds()
	if progress.Status != "FAIL" {
		progress.Status = status
	}
	logger.WriteLog(cfg.LogFile, progress.Status, executionTime, dir)
	progressManager.UpdateProgress(dir, progress)
}
