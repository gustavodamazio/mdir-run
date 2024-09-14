package config

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	InitialDir         string
	Commands           [][]string
	Concurrency        int
	LogFile            string
	SubDirsEntryPoints []string
}

func ParseConfig() (*Config, error) {
	// Define flags
	commandsFlag := flag.String("commands", "", "Commands to execute, separated by semicolons")
	dirFlag := flag.String("dir", "", "Directory in which to execute")
	concurrencyFlag := flag.Int("concurrency", 10, "Number of concurrent operations")
	subDirsFlag := flag.String("subdirs", "", "Subdirectories entry points to run commands in, separated by semicolons")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)

	// Abstracted input parsing
	initialDir := getInput("Enter the directory in which to execute: ", dirFlag, reader)
	commandsInput := getInput("Enter the commands to execute, separated by semicolons: ", commandsFlag, reader)

	// Process subdirectories
	subDirs := []string{}
	if *subDirsFlag != "" {
		for _, subDir := range strings.Split(*subDirsFlag, ";") {
			subDirs = append(subDirs, strings.TrimSpace(subDir))
		}
	}

	// Process commands
	commandsStrings := strings.Split(commandsInput, ";")
	var commands [][]string
	for _, cmd := range commandsStrings {
		cmd = strings.TrimSpace(cmd)
		cmdArgs := strings.Fields(cmd)
		commands = append(commands, cmdArgs)
	}

	// Create log file path
	logFile := filepath.Join(initialDir, "script.log")

	return &Config{
		InitialDir:         initialDir,
		Commands:           commands,
		Concurrency:        *concurrencyFlag,
		LogFile:            logFile,
		SubDirsEntryPoints: subDirs,
	}, nil
}

func getInput(prompt string, flagValue *string, reader *bufio.Reader) string {
	if *flagValue == "" {
		fmt.Print(prompt)
		input, _ := reader.ReadString('\n')
		return strings.TrimSpace(input)
	}
	return *flagValue
}
