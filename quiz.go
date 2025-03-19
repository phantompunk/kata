package main

import (
	"fmt"
	"path/filepath"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	question, err := kata.Questions.GetRandom()
	if err != nil {
		return err
	}
	problem := question.ToProblem(kata.Config.Language)

	fmt.Println("Problem stubbed at", filepath.Join(kata.Config.Workspace, problem.SolutionFilePath()))
	if kata.Config.OpenInEditor {
		openWithEditor(filepath.Join(kata.Config.Workspace, problem.SolutionFilePath()))
	}
	return nil
}
