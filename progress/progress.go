package progress

import (
	"fmt"
	"sync"

	"github.com/gosuri/uilive"
)

type Progress struct {
	Dir      string
	Step     int
	Total    int
	Command  string
	Status   string
	Output   string
	StartRow int
}

type ProgressManager struct {
	mu            sync.Mutex
	progressMap   map[string]*Progress
	progressOrder []string
}

func NewProgressManager(dirs []string) *ProgressManager {
	pm := &ProgressManager{
		progressMap:   make(map[string]*Progress),
		progressOrder: make([]string, 0, len(dirs)),
	}

	for _, dir := range dirs {
		pm.progressMap[dir] = &Progress{
			Dir:      dir,
			Step:     0,
			Total:    0,
			Command:  "Initializing",
			Status:   "Processing",
			Output:   "",
			StartRow: len(pm.progressOrder) + 1,
		}
		pm.progressOrder = append(pm.progressOrder, dir)
	}
	return pm
}

func (pm *ProgressManager) UpdateProgress(dir string, progress *Progress) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.progressMap[dir] = progress
}

func (pm *ProgressManager) GetProgress(dir string) *Progress {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.progressMap[dir]
}

func (pm *ProgressManager) PrintAllProgress(writer *uilive.Writer) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, dir := range pm.progressOrder {
		progress := pm.progressMap[dir]
		if progress.Status == "Processing" {
			if progress.Total > 0 {
				fmt.Fprintf(writer, "%s | step: %d/%d | command: %s\n", progress.Dir, progress.Step, progress.Total, progress.Command)
			} else {
				fmt.Fprintf(writer, "%s | %s\n", progress.Dir, progress.Command)
			}
		} else if progress.Status == "FAIL" {
			fmt.Fprintf(writer, "%s | ERROR: %s\n%s\n", progress.Dir, progress.Command, progress.Output)
		} else {
			fmt.Fprintf(writer, "%s | %s\n", progress.Dir, progress.Status)
		}
	}
	writer.Flush()
}
