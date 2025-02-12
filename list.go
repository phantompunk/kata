package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func ListFunc(cmd *cobra.Command, args []string) error {
	ensureConfig()
	fmt.Println("Reading from", cfg.Workspace)

	// Read from Workspace
	// matches, err := os.ReadDir(filepath.Join(cfg.Workspace, "python"))
	matches, err := filepath.Glob(filepath.Join(cfg.Workspace, "python", "*.py"))
	if err != nil {
		return fmt.Errorf("Glob error: %w", err)
	}

	// List files
	excludePattern := "_test"
	for _, file := range matches {
		fileName := filepath.Base(file)
		if !strings.Contains(fileName, excludePattern) {
			fmt.Println(fileName)
		}
	}
	return nil
}
