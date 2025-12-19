package cmd

import (
	"errors"
	"fmt"
	"time"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/leetcode"
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
		return err
	}

	opts := app.AppOptions{
		Problem:   problemName,
		Language:  language,
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

	submissionId, err := kata.Question.SubmitSolution(cmd.Context(), problem, opts)
	if err != nil {
		return err
	}
	fmt.Print("✔ Submitting solution to leetcode")

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

	displaySubmissionResults(result)
	return nil
}

func displaySubmissionResults(result *leetcode.SubmissionResult) {
	ui.Print("")
	ui.PrintSuccess("Submission accepted!\n")

	ui.Print(fmt.Sprintf("Result:  %s", result.Result))
	ui.Print(fmt.Sprintf("Runtime:  %s (beats %s of Go submissions)", result.Runtime, result.RuntimePercentile))
	ui.Print(fmt.Sprintf("Memory:   %s MB (beats %s)", result.Memory, result.MemoryPercentile))

	ui.Print("\n🎉 Great job! Your solution was accepted.")
}

func displayWaitForResults(start time.Time, wait time.Duration, done <-chan struct{}) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if time.Since(start) >= wait {
				return
			}
			fmt.Print(".")
		}
	}
}
