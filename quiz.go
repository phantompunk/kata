package main

import (
	"fmt"

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
	problem := question.ToProblem(kata.Config.Workspace, kata.Config.Language)

	fmt.Println("Problem stubbed at", problem.SolutionPath)
	if kata.Config.OpenInEditor {
		openWithEditor(problem.SolutionPath)
	}
	return nil
}
