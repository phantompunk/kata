package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func QuizFunc(cmd *cobra.Command, args []string) error {
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
