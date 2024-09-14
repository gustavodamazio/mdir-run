package directories

import (
	"fmt"
	"os"
)

func GetDirectories(initialDir string) ([]string, error) {
	entries, err := os.ReadDir(initialDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", initialDir, err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "." && entry.Name() != ".." {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}
