package gui

import (
	"fmt"
	"image/color"
	"os"
	"runtime"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
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
	app               fyne.App
	window            fyne.Window
	progressItems     map[string]*widget.Label
	progressList      *widget.List
	progressData      []string
	progressColors    map[int]color.Color // Colors for list items
	cfg               *config.Config
	executeButton     *widget.Button
	logArchivePath    string
	statusLine1       *canvas.Text // First line: "Execution completed"
	statusLine2       *canvas.Text // Second line: Success/Failure counts
	statusLine3       *canvas.Text // Third line: Log file path
	statusColor       color.Color  // Current status color
}

// LaunchGUI starts the GUI application
func LaunchGUI() {
	g := &GUI{
		app:            initializeApp(),
		progressItems:  make(map[string]*widget.Label),
		progressData:   []string{},
		progressColors: make(map[int]color.Color),
		statusLine1:    canvas.NewText("", color.White),    // First line, initialized with white color
		statusLine2:    canvas.NewText("", color.White),    // Second line, initialized with white color
		statusLine3:    canvas.NewText("", color.White),    // Third line, initialized with white color 
		statusColor:    color.White,                        // Default status color
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

	// Progress list with colored items
	g.progressList = widget.NewList(
		func() int {
			return len(g.progressData)
		},
		func() fyne.CanvasObject {
			// Create a text object that can be colored
			text := canvas.NewText("Template", color.White)
			text.TextSize = 14
			return container.NewCenter(text)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			fyne.Do(func() {
				// Get the text object from the container
				text := obj.(*fyne.Container).Objects[0].(*canvas.Text)
				text.Text = g.progressData[id]
				
				// Set color if one is assigned
				if c, ok := g.progressColors[id]; ok {
					text.Color = c
				} else {
					text.Color = color.White
				}
				
				text.Refresh()
			})
		},
	)

	// No output text area anymore

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

	// Status summary at the bottom - configure each line
	// First status line - Completion status
	g.statusLine1.Alignment = fyne.TextAlignCenter
	g.statusLine1.TextSize = 20
	g.statusLine1.Text = ""
	
	// Second status line - Success/Failure counts
	g.statusLine2.Alignment = fyne.TextAlignCenter
	g.statusLine2.TextSize = 18
	g.statusLine2.Text = ""
	
	// Third status line - Log file path
	g.statusLine3.Alignment = fyne.TextAlignCenter
	g.statusLine3.TextSize = 16
	g.statusLine3.Text = ""
	
	// Create horizontal scrollers for each line
	statusScroller1 := container.NewHScroll(g.statusLine1)
	statusScroller2 := container.NewHScroll(g.statusLine2)
	statusScroller3 := container.NewHScroll(g.statusLine3)
	
	// Create a padded container with more space for all three lines
	statusContainer := container.NewPadded(
		container.NewVBox(
			widget.NewSeparator(),
			container.NewPadded(statusScroller1),
			container.NewPadded(statusScroller2),
			container.NewPadded(statusScroller3),
		),
	)

	// Main layout with status summary at the bottom
	content := container.NewBorder(
		form,
		statusContainer, // Status at the bottom
		nil,
		nil,
		g.progressList, // Progress list takes the full center area
	)
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

	// Clear progress data and set default colors
	g.progressData = []string{}
	g.progressColors = make(map[int]color.Color)
	g.statusColor = color.White // Reset status color to white
	fyne.Do(func() {
		g.progressList.Refresh()
		// Clear and reset all status lines
		g.statusLine1.Text = ""
		g.statusLine1.Color = color.White
		g.statusLine1.Refresh()
		g.statusLine2.Text = ""
		g.statusLine2.Color = color.White
		g.statusLine2.Refresh()
		g.statusLine3.Text = ""
		g.statusLine3.Color = color.White
		g.statusLine3.Refresh()
	})
	g.logArchivePath = ""

	// Start execution in a goroutine
	go func() {
		g.startExecution()
		// The colored completion status is shown via updateCompletionStatus
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

	// Initialize progress data with white text
	g.progressData = make([]string, len(dirs))
	for i, dir := range dirs {
		g.progressData[i] = fmt.Sprintf("%s | Waiting...", dir)
		g.progressColors[i] = color.White // Ensure text starts as white
	}
	fyne.Do(func() {
		g.progressList.Refresh()
	})

	// Create a custom progress manager that updates the GUI
	progressManager := NewGUIProgressManager(dirs, g)

	// Execute commands
	executor.ExecuteCommands(dirs, g.cfg, &progressManager.ProgressManager)

	// We don't need to forcibly update statuses at the end because
	// the progress manager already properly updates the status
	// via the GUIProgressManager.UpdateProgress method.
	// Any remaining "Waiting..." statuses indicate directories that weren't processed.
	
	// For backward compatibility, update any remaining "Waiting..." entries
	var updatedIndexes []int
	var updatedValues []string

	for i, statusText := range g.progressData {
		if strings.Contains(statusText, "Waiting...") {
			dirName := strings.Split(statusText, " | ")[0]
			// Get the real status from progress manager if available
			progress := progressManager.GetProgress(dirName)
			
			updatedIndexes = append(updatedIndexes, i)
			if progress != nil && progress.Status != "Processing" {
				updatedValues = append(updatedValues, fmt.Sprintf("%s | %s", dirName, progress.Status))
			} else {
				// Fallback if no status is available
				updatedValues = append(updatedValues, fmt.Sprintf("%s | Not processed", dirName))
			}
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
	
	// Update completion status with color
	g.updateCompletionStatus()

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
	// Output is now ignored since we removed the output text area
	// This method is kept for compatibility with existing code
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

// updateCompletionStatus analyzes all progress items and updates the status summary with appropriate color
func (g *GUI) updateCompletionStatus() {
	// Count successes and failures
	successCount := 0
	failCount := 0
	
	// Define colors
	successColor := color.RGBA{0, 180, 0, 255}   // Green
	mixedColor := color.RGBA{255, 140, 0, 255}   // Orange
	failColor := color.RGBA{220, 20, 20, 255}    // Red
	
	// Analyze progress data
	for i, status := range g.progressData {
		if strings.Contains(status, "SUCCESS") {
			successCount++
			// Color successful items green
			g.progressColors[i] = successColor
		} else if strings.Contains(status, "FAIL") {
			failCount++
			// Color failed items red
			g.progressColors[i] = failColor
		}
	}
	
	// Prepare the three separate status lines
	line1Text := "--- Execution completed ---"
	
	// Line 2: Success/failure stats
	line2Text := ""
	totalItems := len(g.progressData)
	if totalItems > 0 {
		line2Text = fmt.Sprintf("Success: %d/%d | Failure: %d/%d", 
			successCount, totalItems, failCount, totalItems)
	}
	
	// Line 3: Log file path
	line3Text := ""
	if g.logArchivePath != "" {
		line3Text = fmt.Sprintf("Log files archived in: %s", g.logArchivePath)
	}
	
	// Determine the appropriate color based on results
	if failCount == 0 && successCount > 0 {
		// All succeeded
		g.statusColor = successColor
	} else if successCount == 0 && failCount > 0 {
		// All failed
		g.statusColor = failColor
	} else if successCount > 0 && failCount > 0 {
		// Mixed results
		g.statusColor = mixedColor
	} else {
		// No results or other case
		g.statusColor = color.White
	}
	
	// Update UI on main thread
	fyne.Do(func() {
		// Update all three status lines with the same color
		g.statusLine1.Text = line1Text
		g.statusLine1.Color = g.statusColor
		g.statusLine1.Refresh()
		
		g.statusLine2.Text = line2Text
		g.statusLine2.Color = g.statusColor
		g.statusLine2.Refresh()
		
		g.statusLine3.Text = line3Text
		g.statusLine3.Color = g.statusColor
		g.statusLine3.Refresh()
		
		// Reset all colors to ensure they display correctly
		for i := range g.progressData {
			if _, ok := g.progressColors[i]; !ok {
				g.progressColors[i] = color.White
			}
		}
		g.progressList.Refresh() // Refresh list to update item colors
	})
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
		status += fmt.Sprintf("%s: %s", progress.Status, progress.Command)
		// We no longer need to update the output text
	} else if progress.Status == "SUCCESS" || strings.HasPrefix(progress.Status, "SUCCESS(") {
		status += fmt.Sprintf("%s", progress.Status)
		// We no longer need to update the output text
	} else {
		status += progress.Status
	}

	pm.gui.updateProgress(dir, status)
}
