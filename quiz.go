package main

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
	ensureConfig()

	var matches []string
	files, err := filepath.Glob(filepath.Join(cfg.Workspace, "python", "*.py"))
	if err != nil {
		return fmt.Errorf("Glob error: %w", err)
	}

	excludePattern := "_test"
	for _, file := range files {
		if !strings.Contains(filepath.Base(file), excludePattern) {
			matches = append(matches, file)
		}
	}

	if len(matches) == 0 {
		fmt.Println("No problems found at", filepath.Join(cfg.Workspace, "python", "*.py"))
		return nil
	}

	randomProblem := matches[rand.Intn(len(matches))]
	randomProblem = strings.ReplaceAll(randomProblem, "_", "-")
	fmt.Println("Try", randomProblem)
	return nil
}
