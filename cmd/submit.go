package cmd

import (
	"errors"
	"time"

	"github.com/phantompunk/kata/internal/app"
	"github.com/phantompunk/kata/internal/ui"
	"github.com/spf13/cobra"
)

var submitCmd = &cobra.Command{
	Use:     "submit",
	Short:   "Submit solutions against leetcode servers",
	PreRunE: validateLanguagePreRun,
	RunE:    HandleErrors(SubmitFunc),
	Args:    cobra.ExactArgs(1),
}

func init() {
	submitCmd.Flags().StringVarP(&language, "language", "l", "", "Programming language to use")
}

func SubmitFunc(cmd *cobra.Command, args []string) error {
	problemName := app.ConvertToSlug(args[0])
	presenter := ui.NewPresenter()

	opts := app.AppOptions{
		Problem:   problemName,
		Language:  language,
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

	submissionId, err := kata.Question.SubmitSolution(cmd.Context(), problem, opts)
	if err != nil {
		return err
	}
	presenter.ShowSubmittingSolution()

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

	presenter.ShowSubmissionResults(result)
	return nil
}

