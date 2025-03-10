package main

import (
	"github.com/phantompunk/kata/internal/config"
	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
	// kata, err := app.New()
	// question, err := kata.questions.GetRandom()
	config.EnsureConfig()

	// var matches []string
	// files, err := filepath.Glob(filepath.Join(cfg.Workspace, "python", "*.py"))
	// if err != nil {
	// 	return fmt.Errorf("Glob error: %w", err)
	// }
	//
	// excludePattern := "_test"
	// for _, file := range files {
	// 	if !strings.Contains(filepath.Base(file), excludePattern) {
	// 		matches = append(matches, file)
	// 	}
	// }
	//
	// if len(matches) == 0 {
	// 	fmt.Println("No problems found at", filepath.Join(cfg.Workspace, "python", "*.py"))
	// 	return nil
	// }
	//
	// randomProblem := matches[rand.Intn(len(matches))]
	// randomProblem = strings.ReplaceAll(randomProblem, "_", "-")
	// fmt.Println("Try", randomProblem)
	return nil
}
