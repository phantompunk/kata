package cmd

import (
	"errors"
	"time"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "Runs problem solution against leetcode test cases",
	PreRunE: validateLanguagePreRun,
	RunE:    HandleErrors(TestFunc),
	Args:    cobra.ExactArgs(1),
}

func init() {
	testCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func TestFunc(cmd *cobra.Command, args []string) error {
	problemName := app.ConvertToSlug(args[0])
	presenter := ui.NewPresenter()

	opts := app.AppOptions{
		Language:  language,
		Problem:   problemName,
		Workspace: kata.Config.WorkspacePath(),
	}

	problem, err := kata.Question.GetBySlug(cmd.Context(), opts)
	if err != nil {
		if errors.Is(err, app.ErrQuestionNotFound) {
			presenter.ShowProblemNotFound(problemName)
			return nil
		}
		return err
	}
	presenter.ShowProblemFetched(problem.Title)

	if !problem.SolutionExists() {
		presenter.ShowSolutionNotFound(problem.Title, problem.Language.DisplayName())
		return nil
	}

	submissionId, err := kata.Question.SubmitTest(cmd.Context(), problem, opts)
	if err != nil {
		return err
	}
	presenter.ShowRunningTests()

	startTime := time.Now()
	maxWait := time.Duration(10) * time.Second

	done := make(chan struct{})
	go presenter.ShowWaitForResults(startTime, maxWait, done)

	result, err := kata.Question.WaitForResult(cmd.Context(), problem, submissionId, maxWait)
	if err != nil {
		if errors.Is(err, app.ErrSolutionFailed) {
			presenter.ShowSolutionFailed()
		}
		return err
	}

	presenter.ShowTestResults(result, problem)
	return nil
}

