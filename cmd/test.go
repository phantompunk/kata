package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/leetcode"
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
		return err
	}

	opts := app.AppOptions{
		Language:  language,
		Problem:   problemName,
		Workspace: kata.Config.WorkspacePath(),
	}

	problem, err := kata.Question.GetBySlug(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrQuestionNotFound) {
			ui.PrintError("Problem %s not found", problemName)
			return nil
		}
		return err
	}
	ui.PrintSuccess("%s", fmt.Sprintf("Fetched problem: %s", problem.Title))

	if !problem.SolutionExists() {
		ui.PrintError("Solution to %q not found using %q", problem.Title, problem.Language.DisplayName())
		return nil
	}

	submissionId, err := kata.Question.SubmitTest(cmd.Context(), problem, opts)
	if err != nil {
		return err
	}
	fmt.Print("✔ Running tests")

	startTime := time.Now()
	maxWait := time.Duration(10) * time.Second

	done := make(chan struct{})
	go displayWaitForResults(startTime, maxWait, done)

	result, err := kata.Question.WaitForResult(cmd.Context(), problem, submissionId, maxWait)
	if err != nil {
		if errors.Is(err, app.ErrSolutionFailed) {
			ui.PrintError("Solution failed")
		}
		return err
	}

	displayTestResults(result, problem)
	return nil
}

func displayTestResults(result *leetcode.SubmissionResult, problem *domain.Problem) {
	if result.HasError() {
		ui.Print("")
		ui.PrintError("Test failed:\n    Error: %q", result.ErrorMsg)
		ui.Print("\nFix your code then try again")
		return
	}

	if !result.IsCorrect() {
		ui.PrintError("\nFailed on test case #%d", result.TestCase)

		ui.Print(fmt.Sprintf("\nInput:    %s", strings.ReplaceAll(problem.Testcases[result.TestCase], "\n", ", ")))
		ui.Print(fmt.Sprintf("Output:   %s", result.TestOutput))
		ui.Print(fmt.Sprintf("Expected: %s", result.TestExpected))
		ui.Print("\nFix your code then try again")
		return
	}

	ui.Print("")
	ui.PrintSuccess("All test cases passed")

	ui.Print("")
	ui.PrintInfo("You are ready to submit")
}
