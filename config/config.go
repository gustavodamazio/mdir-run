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
	Retries            int
}

// Command line flags
var (
	CommandsFlag    = flag.String("commands", "", "Commands to execute, separated by semicolons")
	DirFlag         = flag.String("dir", "", "Directory in which to execute")
	ConcurrencyFlag = flag.Int("concurrency", 10, "Number of concurrent operations")
	SubDirsFlag     = flag.String("subdirs", "", "Subdirectories entry points to run commands in, separated by semicolons")
	RetriesFlag     = flag.Int("retries", 0, "Number of retries for failed commands")
)

func ParseConfig() (*Config, error) {
	// No need to parse flags here now, main.go handles that

	reader := bufio.NewReader(os.Stdin)

	// Abstracted input parsing
	initialDir := getInput("Enter the directory in which to execute: ", DirFlag, reader)
	commandsInput := getInput("Enter the commands to execute, separated by semicolons: ", CommandsFlag, reader)

	// Process subdirectories
	subDirs := []string{}
	if *SubDirsFlag != "" {
		for _, subDir := range strings.Split(*SubDirsFlag, ";") {
			subDirs = append(subDirs, strings.TrimSpace(subDir))
		}
	}

	// Process commands
	commandsStrings := strings.Split(commandsInput, ";")
	var commands [][]string
	for _, cmd := range commandsStrings {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		cmdArgs := strings.Fields(cmd)
		commands = append(commands, cmdArgs)
	}

	// Create log file path
	logFile := filepath.Join(initialDir, "script.log")

	return &Config{
		InitialDir:         initialDir,
		Commands:           commands,
		Concurrency:        *ConcurrencyFlag,
		LogFile:            logFile,
		SubDirsEntryPoints: subDirs,
		Retries:            *RetriesFlag,
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
