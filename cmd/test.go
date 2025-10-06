package cmd

import (
	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Runs problem solution against leetcode test cases",
	RunE:  HandleErrors(TestFunc),
	Args:  cobra.ExactArgs(1),
}

func init() {
	testCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func TestFunc(cmd *cobra.Command, args []string) error {
	problemName := app.ConvertToSlug(args[0])

	if err := validateLanguage(); err != nil {
		ui.PrintError("language %q not supported", language)
		return nil
	}

	opts := app.AppOptions{
		Language:  language,
		Problem:   problemName,
		Workspace: kata.Config.WorkspacePath(),
	}

	// problem, err := kata.Download.GetBySlug(cmd.Context(), opts)
	// if err != nil {
	// 	if errors.Is(err, app.ErrQuestionNotFound) {
	// 		ui.PrintError("Problem %s not found", problemName)
	// 		return nil
	// 	}
	// 	return err
	// }
	// ui.PrintSuccess(fmt.Sprintf("Fetched problem: %s", problem.Title))
	//
	// if !problem.SolutionExists() {
	// 	ui.PrintError("Solution to %q not found using %q", problem.Title, problem.Language.DisplayName())
	// 	return nil
	// }
	//
	// submissionId, err := kata.Download.SubmitTest(cmd.Context(), problem, opts)
	// if err != nil {
	// 	return err
	// }
	// ui.PrintSuccess("Running tests...")

	// ui.PrintInfo("Waiting for " + submissionId)

	// result, err := kata.Download.WaitForResult(cmd.Context(), submissionId, maxWait)
	// if err != nil {
	// 	if errors.Is(err, app.ErrSolutionFailed) {
	// 		ui.PrintError("Solution failed")
	// 	}
	// 	return err
	// }
	// return nil

	return kata.Test(opts)
}
