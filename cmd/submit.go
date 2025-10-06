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

	submissionId, err := kata.Download.SubmitTest(cmd.Context(), problem, opts)
	if err != nil {
		return err
	}
	ui.PrintSuccess("Submitting solution to leetcode")

	// Display wait time til result
	// Print ... while waiting
	startTime := time.Now()
	maxWait := time.Duration(10) * time.Second
	displayWaitForResults(startTime, maxWait)
	//
	// // wait til result
	ui.PrintInfo("Waiting for " + submissionId)
	// result, err := kata.Download.WaitForResult(cmd.Context(), submissionId, maxWait)
	// if err != nil {
	// 	if errors.Is(err, app.ErrSolutionFailed) {
	// 		ui.PrintError("Solution failed")
	// 	}
	// 	return err
	// }
	// // print result summary
	// displaySubmissionResults(result)

	return nil
}

func displaySubmissionResults(result *leetcode.SubmissionResult) {
	ui.PrintSuccess("Submission accepted!\n")

	ui.PrintInfo(fmt.Sprintf("Result:  %s", result.Result))
	ui.PrintInfo(fmt.Sprintf("Runtime:  %s (beats %s of Go submissions)", result.Runtime, result.RuntimeMsg))
	ui.PrintInfo(fmt.Sprintf("Memory:   %s MB (beats %s)", result.Memory, result.MemoryMsg))

	ui.PrintInfo("🎉 Great job! Your solution was accepted.")
}

func displayWaitForResults(start time.Time, wait time.Duration) {
	go func() {
		for {
			elapsed := time.Since(start)
			if elapsed >= wait {
				break
			}
			fmt.Print(".")
			time.Sleep(1 * time.Second)
		}
	}()
}
