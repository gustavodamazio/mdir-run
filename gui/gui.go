package gui

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/gustavodamazio/mdir-run/config"
	"github.com/gustavodamazio/mdir-run/directories"
	"github.com/gustavodamazio/mdir-run/executor"
	"github.com/gustavodamazio/mdir-run/logger"
	"github.com/gustavodamazio/mdir-run/progress"
)

// Custom layout to enforce a fixed width
type fixedWidthLayout struct {
	width float32
}

func (f *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := f.width
	h := float32(0)
	for _, o := range objects {
		if o.Visible() {
			min := o.MinSize()
			if min.Height > h {
				h = min.Height
			}
		}
	}
	return fyne.NewSize(w, h)
}

func (f *fixedWidthLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, o := range objects {
		if o.Visible() {
			min := o.MinSize()
			o.Resize(fyne.NewSize(f.width, min.Height))
			o.Move(fyne.NewPos(0, (containerSize.Height-min.Height)/2))
		}
	}
}

type GUI struct {
	app            fyne.App
	window         fyne.Window
	progressItems  map[string]*widget.Label
	progressList   *widget.List
	outputText     *widget.Entry
	progressData   []string
	cfg            *config.Config
	executeButton  *widget.Button
	logArchivePath string
}

// LaunchGUI starts the GUI application
func LaunchGUI() {
	g := &GUI{
		app:           initializeApp(),
		progressItems: make(map[string]*widget.Label),
		progressData:  []string{},
		cfg: &config.Config{
			Concurrency:        10,
			Retries:            2,
			SubDirsEntryPoints: []string{"functions"},
		},
	}

	g.window = g.app.NewWindow("mdir-run")
	g.window.Resize(fyne.NewSize(800, 600))
	g.buildUI()
	g.window.ShowAndRun()
}

func initializeApp() fyne.App {
	if isRunningOnMacOS() {
		return app.NewWithID("com.gustavodamazio.mdir-run") // Use specific app ID for macOS
	}
	return app.New()
}

func isRunningOnMacOS() bool {
	return runtime.GOOS == "darwin"
}

func (g *GUI) buildUI() {
	// Directory input field
	dirEntry := widget.NewEntry()
	dirEntry.SetPlaceHolder("Directory path...")

	// Browse button
	browseButton := widget.NewButtonWithIcon("Browse", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil || uri == nil {
				return
			}
			dirEntry.SetText(uri.Path())
		}, g.window)
	})

	// Commands input field
	commandsEntry := widget.NewMultiLineEntry()
	commandsEntry.SetPlaceHolder("Enter commands (one command per line)")

	// Subdirectories input field
	subdirsEntry := widget.NewEntry()
	subdirsEntry.SetText(strings.Join(g.cfg.SubDirsEntryPoints, ";"))
	subdirsEntry.SetPlaceHolder("Optional subdirectories")

	// Concurrency input
	concurrencyEntry := widget.NewEntry()
	concurrencyEntry.SetText(fmt.Sprintf("%d", g.cfg.Concurrency))

	// Retries input
	retriesEntry := widget.NewEntry()
	retriesEntry.SetText(fmt.Sprintf("%d", g.cfg.Retries))

	// Execute button
	g.executeButton = widget.NewButtonWithIcon("Execute", theme.MediaPlayIcon(), func() {
		// Disable button during execution
		g.executeButton.Disable()
		g.executeCommands(dirEntry.Text, commandsEntry.Text, subdirsEntry.Text, concurrencyEntry.Text, retriesEntry.Text)
		// Re-enable button when execution completes (done in startExecution)
	})

	// Progress list
	g.progressList = widget.NewList(
		func() int {
			return len(g.progressData)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			fyne.Do(func() {
				obj.(*widget.Label).SetText(g.progressData[id])
			})
		},
	)

	// Output text area
	g.outputText = widget.NewMultiLineEntry()
	g.outputText.Disable()

	// Create labels with consistent width
	dirLabel := widget.NewLabel("Directory:")
	cmdLabel := widget.NewLabel("Commands:")
	subdirsLabel := widget.NewLabel("Subdirectories:")
	concurrencyLabel := widget.NewLabel("Concurrency:")
	retriesLabel := widget.NewLabel("Retries:")

	// Fixed width for labels
	dirLabelContainer := container.New(&fixedWidthLayout{width: 100}, dirLabel)
	cmdLabelContainer := container.New(&fixedWidthLayout{width: 100}, cmdLabel)
	subdirsLabelContainer := container.New(&fixedWidthLayout{width: 100}, subdirsLabel)
	concurrencyLabelContainer := container.New(&fixedWidthLayout{width: 100}, concurrencyLabel)
	retriesLabelContainer := container.New(&fixedWidthLayout{width: 100}, retriesLabel)

	// Form layout using fixed-width labels and full width entries
	form := container.NewVBox(
		container.NewBorder(nil, nil, dirLabelContainer, browseButton, dirEntry),
		container.NewBorder(nil, nil, cmdLabelContainer, nil, commandsEntry),
		container.NewBorder(nil, nil, subdirsLabelContainer, nil, subdirsEntry),
		container.NewHBox(concurrencyLabelContainer, container.New(&fixedWidthLayout{width: 100}, concurrencyEntry)),
		container.NewHBox(retriesLabelContainer, container.New(&fixedWidthLayout{width: 100}, retriesEntry)),
		g.executeButton,
	)

	// Main split layout
	split := container.NewHSplit(
		g.progressList,
		g.outputText,
	)
	split.SetOffset(0.3) // 30% for progress list, 70% for output

	// Main layout
	content := container.NewBorder(form, nil, nil, nil, split)
	g.window.SetContent(content)
}

func (g *GUI) executeCommands(dirPath, commandsText, subdirs, concurrency, retries string) {
	// Validate inputs
	if dirPath == "" {
		dialog.ShowError(fmt.Errorf("directory path cannot be empty"), g.window)
		// Re-enable button
		fyne.Do(func() {
			g.executeButton.Enable()
		})
		return
	}

	if commandsText == "" {
		dialog.ShowError(fmt.Errorf("commands cannot be empty"), g.window)
		// Re-enable button
		fyne.Do(func() {
			g.executeButton.Enable()
		})
		return
	}

	// Process commands from GUI (convert newlines to semicolons)
	commandsStr := strings.ReplaceAll(commandsText, "\n", ";")

	// Convert directory to absolute path if needed
	if !strings.HasPrefix(dirPath, "/") {
		currentDir, err := os.Getwd()
		if err == nil {
			dirPath = currentDir + "/" + dirPath
		}
	}

	// Setup config
	g.cfg.InitialDir = dirPath
	g.cfg.LogFile = dirPath + "/script.log"

	// Process concurrency
	if concurrency != "" {
		fmt.Sscanf(concurrency, "%d", &g.cfg.Concurrency)
	}

	// Process retries
	if retries != "" {
		fmt.Sscanf(retries, "%d", &g.cfg.Retries)
	}

	// Process commands
	commandsStrings := strings.Split(commandsStr, ";")
	var commands [][]string
	for _, cmd := range commandsStrings {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		cmdArgs := strings.Fields(cmd)
		commands = append(commands, cmdArgs)
	}
	g.cfg.Commands = commands

	// Process subdirectories
	if subdirs != "" {
		g.cfg.SubDirsEntryPoints = []string{}
		for _, subDir := range strings.Split(subdirs, ";") {
			g.cfg.SubDirsEntryPoints = append(g.cfg.SubDirsEntryPoints, strings.TrimSpace(subDir))
		}
	}

	// Clear progress data
	g.progressData = []string{}
	fyne.Do(func() {
		g.progressList.Refresh()
		g.outputText.SetText("")
	})
	g.logArchivePath = ""

	// Start execution in a goroutine
	go func() {
		g.startExecution()
		// Update UI when execution is complete
		completionMessage := "\n--- Execution completed ---\n"
		if g.logArchivePath != "" {
			completionMessage += fmt.Sprintf("Log files archived to: %s\n", g.logArchivePath)
		}
		g.updateOutput(completionMessage)
	}()
}

func (g *GUI) startExecution() {
	startTime := time.Now()

	// Initialize log file
	err := logger.InitializeLogFile(g.cfg.LogFile)
	if err != nil {
		g.updateOutput(fmt.Sprintf("Failed to initialize log file: %v\n", err))
		// Re-enable execute button on main thread
		fyne.Do(func() {
			g.executeButton.Enable()
		})
		return
	}

	// Get directories
	dirs, err := directories.GetDirectories(g.cfg.InitialDir)
	if err != nil {
		g.updateOutput(fmt.Sprintf("Failed to get directories: %v\n", err))
		// Re-enable execute button on main thread
		fyne.Do(func() {
			g.executeButton.Enable()
		})
		return
	}

	// Initialize progress data
	g.progressData = make([]string, len(dirs))
	for i, dir := range dirs {
		g.progressData[i] = fmt.Sprintf("%s | Waiting...", dir)
	}
	fyne.Do(func() {
		g.progressList.Refresh()
	})

	// Create a custom progress manager that updates the GUI
	progressManager := NewGUIProgressManager(dirs, g)

	// Execute commands
	executor.ExecuteCommands(dirs, g.cfg, &progressManager.ProgressManager)

	// Prepare updates to status
	var updatedIndexes []int
	var updatedValues []string

	for i, statusText := range g.progressData {
		if strings.Contains(statusText, "Waiting...") {
			dirName := strings.Split(statusText, " | ")[0]
			updatedIndexes = append(updatedIndexes, i)
			updatedValues = append(updatedValues, fmt.Sprintf("%s | Completed", dirName))
		}
	}

	// Apply updates in the main thread
	if len(updatedIndexes) > 0 {
		fyne.Do(func() {
			for j, idx := range updatedIndexes {
				g.progressData[idx] = updatedValues[j]
			}
			g.progressList.Refresh()
		})
	}

	// Write summary log
	logger.WriteSummaryLog(g.cfg.LogFile, startTime)

	// Archive logs
	var archivePath string
	if archivePath, err = logger.ArchiveLogs(g.cfg.LogFile); err != nil {
		g.updateOutput(fmt.Sprintf("WARNING: Failed to archive log files: %v\n", err))
	} else {
		g.logArchivePath = archivePath
	}

	// Re-enable execute button
	fyne.Do(func() {
		g.executeButton.Enable()
	})
}

func (g *GUI) updateProgress(dir string, status string) {
	// UI updates using fyne.Do to ensure the use of the main thread
	fyne.Do(func() {
		// Find the directory in progress data
		for i, d := range g.progressData {
			if strings.HasPrefix(d, dir+" |") {
				g.progressData[i] = status
				break
			}
		}
		g.progressList.Refresh()
	})
}

func (g *GUI) updateOutput(text string) {
	// We use the fyne.Do method to ensure that UI updates
	// happen on Fyne's main thread
	fyne.Do(func() {
		currentText := g.outputText.Text
		g.outputText.SetText(currentText + text)
		// Auto-scroll to bottom
		g.outputText.CursorRow = len(strings.Split(g.outputText.Text, "\n")) - 1
	})
}

// Custom progress manager for GUI
type GUIProgressManager struct {
	progress.ProgressManager
	gui *GUI
}

func NewGUIProgressManager(dirs []string, gui *GUI) *GUIProgressManager {
	pm := &GUIProgressManager{
		ProgressManager: *progress.NewProgressManager(dirs),
		gui:             gui,
	}
	return pm
}

func (pm *GUIProgressManager) UpdateProgress(dir string, progress *progress.Progress) {
	// Update underlying progress manager
	pm.ProgressManager.UpdateProgress(dir, progress)

	// Update GUI
	status := fmt.Sprintf("%s | ", dir)
	if progress.Status == "Processing" {
		if progress.Total > 0 {
			status += fmt.Sprintf("step: %d/%d | command: %s", progress.Step, progress.Total, progress.Command)
		} else {
			status += progress.Command
		}
	} else if progress.Status == "FAIL" || strings.HasPrefix(progress.Status, "FAIL(") {
		status += fmt.Sprintf("ERROR: %s", progress.Command)
		pm.gui.updateOutput(fmt.Sprintf("[%s] ERROR: %s\n%s\n", dir, progress.Command, progress.Output))
	} else if progress.Status == "SUCCESS" || strings.HasPrefix(progress.Status, "SUCCESS(") {
		status += "Completed successfully"
		pm.gui.updateOutput(fmt.Sprintf("[%s] Completed successfully\n", dir))
	} else {
		status += progress.Status
	}

	pm.gui.updateProgress(dir, status)
}
