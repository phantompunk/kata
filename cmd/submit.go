package cmd

import (
	"errors"
	"fmt"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit solutions against leetcode servers",
	RunE:  HandleErrors(SubmitFunc),
	Args:  cobra.ExactArgs(1),
}

func init() {
	submitCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func SubmitFunc(cmd *cobra.Command, args []string) error {
	problemName := app.ConvertToSlug(args[0])

	if err := validateLanguage(); err != nil {
		ui.PrintError("language %q not supported", language)
		return nil
	}

	opts := app.AppOptions{
		Problem:   problemName,
		Language:  language,
		Workspace: kata.Config.WorkspacePath(),
	}

	problem, err := kata.Download.GetBySlug(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrQuestionNotFound) {
			ui.PrintError("Problem %s not found", problemName)
			return nil
		}
		return err
	}
	ui.PrintSuccess(fmt.Sprintf("Fetched problem: %s", problem.Title))

	if !problem.SolutionExists() {
		ui.PrintError("Solution to %q not found using %q", problem.Title, problem.Language.DisplayName())
		return nil
	}

	kata.Download.SubmitQuestion(cmd.Context(), problem, opts)

	return kata.Submit(opts)
}
