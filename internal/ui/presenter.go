package ui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/andanhm/go-prettytime"
	"github.com/dustin/go-humanize"
	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/render"
	"github.com/phantompunk/kata/internal/repository"
)

// Presenter handles all UI output and formatting
type Presenter struct {
	writer io.Writer
}

// Default presenter for package-level operations
var defaultPresenter = &Presenter{writer: os.Stdout}

// Template definitions
const (
	quizTemplate = `
Problem: {{.Title}}
Difficulty: {{.Difficulty}}
Last attempted: {{ prettytime .LastAttempted}}
Status: {{.Status}}

Next steps:
  • Start solving: kata solve {{.Slug}}
  • View details: kata show {{.Slug}}
  • Submit later: kata submit {{.Slug}}
`

	loginTemplate = `
Account:	{{.Username}}
Problems:	{{.Attempted}} attempted, {{.Completed}} completed

You're all set! 🎉

Next steps:
  • Stub problem:     kata get two-sum
  • Test solution:    kata test two-sum
  • Submit solution:  kata submit two-sum
`
)

// NewPresenter creates a new Presenter instance
func NewPresenter() *Presenter {
	return &Presenter{
		writer: os.Stdout,
	}
}

// NewPresenterWithWriter creates a new Presenter with a custom writer (useful for testing)
func NewPresenterWithWriter(w io.Writer) *Presenter {
	return &Presenter{
		writer: w,
	}
}

// ShowWarnings displays a list of warning messages
func ShowWarnings(warnings []string) {
	for _, warning := range warnings {
		defaultPresenter.error("%s", warning)
	}
}

// Internal output methods that use p.writer

func (p *Presenter) success(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintf(p.writer, "✔ %s\n", msg)
}

func (p *Presenter) info(message string) {
	_, _ = fmt.Fprintf(p.writer, "ℹ %s\n", message)
}

func (p *Presenter) error(format string, a ...any) {
	msg := fmt.Sprintf(format, a...)
	_, _ = fmt.Fprintf(p.writer, "✘ %s\n", msg)
}

func (p *Presenter) print(message string) {
	_, _ = fmt.Fprintln(p.writer, message)
}

func (p *Presenter) warning(message string) {
	_, _ = fmt.Fprintf(p.writer, "⚠ %s\n", message)
}

func (p *Presenter) nextSteps(slug string) {
	tmpl := `
Next steps:
  • Test solution:    kata test %s
  • Submit solution:  kata submit %s
`
	_, _ = fmt.Fprintf(p.writer, tmpl, slug, slug)
}

// ShowQuizResult displays the quiz result using a template
func (p *Presenter) ShowQuizResult(problem *domain.Problem) error {
	p.success("Selected a random problem from your history")
	return p.renderQuizResult(problem)
}

// ShowTestResults displays test execution results
func (p *Presenter) ShowTestResults(result *leetcode.SubmissionResult, problem *domain.Problem) {
	if result.HasError() {
		p.print("")
		p.error("Test failed:\n    Error: %q", result.ErrorMsg)
		p.print("\nFix your code then try again")
		return
	}

	if !result.IsCorrect() {
		p.error("\nFailed on test case #%d", result.TestCase)

		p.print(fmt.Sprintf("\nInput:    %s", strings.ReplaceAll(problem.Testcases[result.TestCase], "\n", ", ")))
		p.print(fmt.Sprintf("Output:   %s", result.TestOutput))
		p.print(fmt.Sprintf("Expected: %s", result.TestExpected))
		p.print("\nFix your code then try again")
		return
	}

	p.print("")
	p.success("All test cases passed")

	p.print("")
	p.info("You are ready to submit")
}

// ShowSubmissionResults displays submission results
func (p *Presenter) ShowSubmissionResults(result *leetcode.SubmissionResult) {
	p.print("")
	p.success("Submission accepted!\n")

	p.print(fmt.Sprintf("Result:  %s", result.Result))
	p.print(fmt.Sprintf("Runtime:  %s (beats %s of Go submissions)", result.Runtime, result.RuntimePercentile))
	p.print(fmt.Sprintf("Memory:   %s MB (beats %s)", result.Memory, result.MemoryPercentile))

	p.print("\n🎉 Great job! Your solution was accepted.")
}

// ShowWaitForResults displays a progress indicator while waiting for results
func (p *Presenter) ShowWaitForResults(start time.Time, wait time.Duration, done <-chan struct{}) {
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
			_, _ = fmt.Fprint(p.writer, ".")
		}
	}
}

// ShowProblemFetched displays a success message for fetched problems
func (p *Presenter) ShowProblemFetched(title string) {
	p.success("Fetched problem: %s", title)
}

// ShowProblemNotFound displays an error when a problem is not found
func (p *Presenter) ShowProblemNotFound(problemName string) {
	p.error("Problem %s not found", problemName)
}

// ShowSolutionNotFound displays an error when a solution file doesn't exist
func (p *Presenter) ShowSolutionNotFound(title, languageName string) {
	p.error("Solution to %q not found using %q", title, languageName)
}

// ShowNoEligibleProblems displays a helpful message when no problems are available for quiz
func (p *Presenter) ShowNoEligibleProblems() {
	p.error("No eligible problems to quiz on")
	p.info("You need at least one attempted solution\n    To get started, run: 'kata get two-sum'")
}

// ShowLanguageNotSupported displays an error for unsupported languages
func (p *Presenter) ShowLanguageNotSupported(language string) {
	p.error("language %q not supported", language)
}

// ShowRunningTests displays a message indicating tests are running
func (p *Presenter) ShowRunningTests() {
	_, _ = fmt.Fprint(p.writer, "✔ Running tests")
}

// ShowSubmittingSolution displays a message indicating solution is being submitted
func (p *Presenter) ShowSubmittingSolution() {
	_, _ = fmt.Fprint(p.writer, "✔ Submitting solution to leetcode")
}

// ShowSolutionFailed displays an error when a solution fails
func (p *Presenter) ShowSolutionFailed() {
	p.error("Solution failed")
}

// ShowProblemAlreadyExists displays a message when a problem already exists
func (p *Presenter) ShowProblemAlreadyExists(title, path, slug string) {
	p.error("Problem %s already exists at:\n  %s", title, path)
	p.print("\nTo retry solution, run:\n  kata get " + slug + " --retry")
	p.print("\nTo force refresh files, run:\n  kata get " + slug + " --force")
}

// ShowProblemDoesNotExist displays a message when attempting to retry a non-existent problem
func (p *Presenter) ShowProblemDoesNotExist(title, path, slug string) {
	p.error("Problem %s does not exist at:\n  %s", title, path)
	p.print("\nTo download the problem first, run:\n  kata get " + slug)
}

// ShowAlreadyLoggedIn displays a message when user is already logged in
func (p *Presenter) ShowAlreadyLoggedIn(username string) {
	p.info("You are already logged in as " + username)
}

// ShowAuthenticationSuccess displays a success message after successful authentication
func (p *Presenter) ShowAuthenticationSuccess() {
	p.success("Authentication successful")
}

// ShowOpeningConfigFile displays a message when opening the config file
func (p *Presenter) ShowOpeningConfigFile(path string) {
	p.success("Opening config file: %s", path)
}

// ShowLoginResult displays the login result with user stats
func (p *Presenter) ShowLoginResult(username string, stats repository.GetStatsRow) error {
	return p.renderLoginResult(username, stats)
}

// ShowRenderResults displays the results of rendering/stubbing a problem
func (p *Presenter) ShowRenderResults(result *render.RenderResult, slug string, force bool) {
	if result.DirectoryCreated != "" {
		p.success("Created directory: %s", result.DirectoryCreated)
	}

	if len(result.FilesCreated) > 0 {
		p.success("Generated files:")
		for _, file := range result.FilesCreated {
			p.print(fmt.Sprintf("  • %s", file))
		}
	}

	if len(result.FilesUpdated) > 0 {
		p.info("Updated files:")
		for _, file := range result.FilesUpdated {
			p.print(fmt.Sprintf("  • %s", file))
		}
	}

	if len(result.FilesSkipped) == 1 && result.FilesSkipped[0] == "All files" {
		p.info(fmt.Sprintf("Problem %s already exists\n", slug))
		if !force {
			p.print(fmt.Sprintf("To refresh files, run:\n  kata get %s --force\n", slug))
		}
		return
	}

	if result.TestSkipped {
		p.warning("Note: Test file generation is not supported for this language")
	}

	if len(result.FilesSkipped) > 0 {
		p.warning("Skipped files:")
		for _, file := range result.FilesSkipped {
			p.info(fmt.Sprintf("  • %s", file))
		}
		if !force {
			p.info("Use --force to overwrite existing files")
		}
	}

	p.nextSteps(slug)
}

// Template rendering methods

func (p *Presenter) renderQuizResult(problem *domain.Problem) error {
	t := template.Must(template.New("Quiz").Funcs(p.templateFuncs()).Parse(quizTemplate))
	return t.Execute(p.writer, problem)
}

func (p *Presenter) renderLoginResult(username string, stats repository.GetStatsRow) error {
	t := template.Must(template.New("Login").Parse(loginTemplate))
	return t.Execute(p.writer, map[string]string{
		"Attempted": fmt.Sprint(stats.Attempted),
		"Completed": fmt.Sprint(stats.Completed),
		"Username":  username,
	})
}

func (p *Presenter) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"humanize":   humanize.Time,
		"prettytime": prettytime.Format,
	}
}
