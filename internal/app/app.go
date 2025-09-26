package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/phantompunk/kata/internal/browser"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/db"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/render/templates"
	"github.com/phantompunk/kata/internal/repository"
	"github.com/spf13/afero"
)

const (
	LoginUrl = "https://leetcode.com/accounts/login/"
)

var (
	ErrCookiesNotFound  = errors.New("session cookies not found")
	ErrNotAuthenticated = errors.New("not authenticated")
	ErrInvalidSession   = errors.New("session is not valid")
	ErrDuplicateProblem = errors.New("question has already been downloaded")
	ErrNoQuestions      = errors.New("no questions found in the database")
)

type AppOptions struct {
	Problem  string
	Language string
	Open     bool
	Force    bool
}

type App struct {
	Config        *config.Config
	QuestionModel *repository.Queries
	Repo          *repository.Queries
	lcs           *leetcode.Service
	Renderer      *templates.Renderer
	fs            afero.Fs
}

func New() (*App, error) {
	cfg, err := config.EnsureConfig()
	if err != nil {
		fmt.Println("Failed cfg")
		return nil, err
	}

	db, err := db.EnsureDB()
	if err != nil {
		fmt.Println("Failed db")
		return nil, err
	}

	repo := repository.New(db)

	lcs, err := leetcode.New(leetcode.WithCookies(cfg.SessionToken, cfg.CsrfToken))
	if err != nil {
		return nil, fmt.Errorf("failed to create leetcode service: %w", err)
	}

	renderer, err := templates.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	return &App{
		Config:   &cfg,
		Repo:     repo,
		lcs:      lcs,
		Renderer: renderer,
		fs:       afero.NewOsFs(),
	}, nil
}

func ConvertToSlug(name string) string {
	if id, err := strconv.Atoi(name); err == nil {
		return MapIDtoSlug[id]
	}

	return name
}

func (app *App) Test(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	opts.Problem = ConvertToSlug(opts.Problem)

	if err := app.CheckSession(); err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	return app.TestSolution(opts.Problem, opts.Language)
}

func (app *App) Submit(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	opts.Problem = ConvertToSlug(opts.Problem)

	if err := app.CheckSession(); err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	return app.SubmitSolution(opts.Problem, opts.Language)
}

// CheckSession checks if the session is valid by pinging the leetcode service
func (app *App) CheckSession() error {
	if !app.Config.IsSessionValid() {
		app.Config.ClearSession()
		return ErrInvalidSession
	}

	valid, err := app.lcs.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping leetcode service: %w", err)
	}

	if !valid {
		app.Config.ClearSession()
		return ErrInvalidSession
	}
	return nil
}

func (app *App) ValidateSession() error {
	username, err := app.lcs.GetUsername()
	if err != nil {
		return fmt.Errorf("failed to ping leetcode service: %w", err)
	}

	if username == "" {
		app.Config.ClearSession()
		return ErrInvalidSession
	}

	return app.Config.SaveUsername(username)
}

// :TODO: Move this to a separate package
func toModelQuestion(repoQuestion repository.Question) (*models.Question, error) {
	var modelQuestion models.Question
	modelQuestion.ID = fmt.Sprintf("%d", repoQuestion.QuestionID)
	modelQuestion.Title = repoQuestion.Title
	modelQuestion.TitleSlug = repoQuestion.TitleSlug
	modelQuestion.Difficulty = repoQuestion.Difficulty
	modelQuestion.FunctionName = repoQuestion.FunctionName
	modelQuestion.Content = repoQuestion.Content

	if err := json.Unmarshal([]byte(repoQuestion.CodeSnippets), &modelQuestion.CodeSnippets); err != nil {
		return nil, err
	}
	modelQuestion.Testcases = repoQuestion.TestCases
	return &modelQuestion, nil
}

func ToProblem(repoQuestion repository.Question, workspace, language string) *models.Problem {
	var problem models.Problem
	problem.QuestionID = fmt.Sprintf("%d", repoQuestion.QuestionID)
	problem.TitleSlug = repoQuestion.TitleSlug
	problem.FunctionName = repoQuestion.FunctionName
	problem.Content = repoQuestion.Content

	var codeSnippets []models.CodeSnippet
	if err := json.Unmarshal([]byte(repoQuestion.CodeSnippets), &codeSnippets); err != nil {
		fmt.Println("Failed to unmarshal code snippets:", err)
		return nil
	}
	problem.Code = ""
	for _, snippet := range codeSnippets {
		if snippet.LangSlug == language {
			problem.Code = snippet.Code
			break
		}
	}
	problem.LangSlug = models.LangName[language]
	problem.SetPaths(workspace)
	return &problem
}

// TODO: Move this to a separate package
func toRepoCreateParams(modelQuestion *models.Question) (repository.CreateParams, error) {
	var params repository.CreateParams
	qId, _ := strconv.ParseInt(modelQuestion.ID, 10, 64)
	params.QuestionID = qId
	params.Title = modelQuestion.Title
	params.TitleSlug = modelQuestion.TitleSlug
	params.Difficulty = modelQuestion.Difficulty
	params.FunctionName = modelQuestion.FunctionName
	params.Content = modelQuestion.Content

	codeSnippets, err := json.Marshal(modelQuestion.CodeSnippets)
	if err != nil {
		return params, err
	}
	params.CodeSnippets = string(codeSnippets)

	testCases := strings.Join(modelQuestion.TestCaseList, "\n")
	params.TestCases = string(testCases)

	return params, nil
}

func (app *App) GetQuestion(name, language string, force bool) (*models.Question, error) {
	exists, err := app.Repo.Exists(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 1 {
		if !force {
			return nil, ErrDuplicateProblem
		}

		repoQuestion, err := app.Repo.GetBySlug(context.Background(), name)
		if err != nil {
			return nil, fmt.Errorf("failed to get question details: %w", err)
		}

		modelQuestion, err := toModelQuestion(repoQuestion)
		if err != nil {
			return nil, fmt.Errorf("failed to convert question: %w", err)
		}

		fmt.Printf("✔ Fetched problem: %s\n", repoQuestion.Title)
		return modelQuestion, nil
	}

	question, err := app.lcs.Fetch(name)
	if err != nil {
		return nil, err
	}

	question.FunctionName = app.GetFunctionName(question.ToProblem(app.Config.Workspace, language))
	params, err := toRepoCreateParams(question)
	if err != nil {
		return nil, fmt.Errorf("failed to convert question to repository params: %w", err)
	}

	app.Repo.Create(context.Background(), params)
	fmt.Printf("✔ Fetched problem: %s\n", question.Title)
	return question, nil
}

func (app *App) Stub(question *models.Question, opts AppOptions) error {
	problem := question.ToProblem(app.Config.Workspace, opts.Language)
	if err := app.fs.MkdirAll(problem.DirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed creating problem directory: %w", err)
	}
	fmt.Printf("✔ Created directory: %s\n", FormatPathForDisplay(problem.DirPath))

	fmt.Println("✔ Generated files:")
	file, err := app.fs.Create(problem.SolutionPath)
	if err != nil {
		return fmt.Errorf("failed creating problem solution file: %w", err)
	}

	test, err := app.fs.Create(problem.TestPath)
	if err != nil {
		return fmt.Errorf("failed create problem test file: %w", err)
	}

	readme, err := app.fs.Create(problem.ReadmePath)
	if err != nil {
		return fmt.Errorf("failed creating readme file: %w", err)
	}

	if err := app.Renderer.RenderFile(file, templates.Solution, problem); err != nil {
		return fmt.Errorf("failed to render solution file: %w", err)
	}
	fmt.Printf("  • %s\n", filepath.Base(problem.SolutionPath))

	if err := app.Renderer.RenderFile(test, templates.Test, problem); err != nil {
		return fmt.Errorf("failed to render test file: %w", err)
	}
	fmt.Printf("  • %s\n", filepath.Base(problem.TestPath))

	if err := app.Renderer.RenderFile(readme, templates.Readme, problem); err != nil {
		return fmt.Errorf("failed to render readme file: %w", err)
	}
	fmt.Printf("  • %s\n", filepath.Base(problem.ReadmePath))

	return nil
}

// TestSolution tests the solution for a given problem name and language.
func (app *App) TestSolution(name, language string) error {
	exists, err := app.Repo.Exists(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("question %q not found", name)
	}

	repoQuestion, err := app.Repo.GetBySlug(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to get question details: %w", err)
	}

	problem := repoQuestion.ToProblem(app.Config.Workspace, language)
	snippet := app.extractSnippet(problem.SolutionPath)

	testStatusUrl, err := app.lcs.Test(problem, language, snippet)
	if err != nil {
		return fmt.Errorf("failed to submit solution for testing: %w", err)
	}

	if testStatusUrl == "" {
		return fmt.Errorf("empty testStatusUrl received from server")
	}

	res, err := app.lcs.PollTestStatus(testStatusUrl)
	if err != nil {
		return fmt.Errorf("failed to poll test status: %w", err)
	}

	if res.Correct {
		fmt.Println()
		fmt.Println("✓ All test cases passed")
		if err := app.Renderer.RenderOutput(os.Stdout, templates.CliTest, problem); err != nil {
			return fmt.Errorf("failed to render quiz: %w", err)
		}
		return nil
	}

	fmt.Println()
	fmt.Println("✗ Some test cases failed")
	if err := app.Renderer.RenderOutput(os.Stdout, templates.CliTest, problem); err != nil {
		return fmt.Errorf("failed to render quiz: %w", err)
	}
	return nil
}

// SubmitSolution tests the solution for a given problem name and language.
func (app *App) SubmitSolution(name, language string) error {
	exists, err := app.Repo.Exists(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 0 {
		return fmt.Errorf("question %q not found", name)
	}

	repoQuestion, err := app.Repo.GetBySlug(context.Background(), name)
	if err != nil {
		return fmt.Errorf("failed to get question details: %w", err)
	}

	problem := repoQuestion.ToProblem(app.Config.Workspace, language)
	snippet := app.extractSnippet(problem.SolutionPath)

	testStatusUrl, err := app.lcs.Submit(problem, language, snippet)
	if err != nil {
		return fmt.Errorf("failed to submit solution for testing: %w", err)
	}

	if testStatusUrl == "" {
		return fmt.Errorf("empty testStatusUrl received from server")
	}

	res, err := app.lcs.PollTestStatus(testStatusUrl)
	if err != nil {
		return fmt.Errorf("failed to poll test status: %w", err)
	}

	if res.Correct || res.StatusMsg == "Accepted" {
		fmt.Println()
		fmt.Println("✓ Submission accepted!")

		if err := app.Renderer.RenderOutput(os.Stdout, templates.CliSubmit, res); err != nil {
			return fmt.Errorf("failed to render quiz: %w", err)
		}

		app.Repo.Submit(context.Background(), repository.SubmitParams{
			QuestionID: repoQuestion.QuestionID,
			LangSlug:   language,
			Solved:     1,
		})
		return nil
	}

	fmt.Println("Failed")
	return fmt.Errorf("Failed: %s.", res.StatusMsg)
}

// TODO: Move this to a separate package
func (app *App) extractSnippet(path string) string {
	file, _ := os.Open(path)
	defer file.Close()

	startMarker := fmt.Sprint("// ::KATA START::")
	endMarker := fmt.Sprint("// ::KATA END::")

	var builder strings.Builder
	scanner := bufio.NewScanner(file)

	inSnippet := false
	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == startMarker {
			inSnippet = true
			continue
		}

		if strings.TrimSpace(line) == endMarker {
			break
		}

		if inSnippet {
			builder.WriteString(line + "\n")
		}
	}
	if err := scanner.Err(); err != nil {
		// todo return error
		return ""
	}
	return strings.TrimSpace(builder.String())
}

func (app *App) GetFunctionName(problem *models.Problem) string {
	var buf bytes.Buffer

	app.Renderer.RenderFile(&buf, templates.Solution, problem)
	name, err := parseFunctionName(buf.String())
	if err != nil {
		fmt.Println("failed %w", err)
		return ""
	}
	return name
}

func (app *App) PrintQuestionStatus() ([]models.Question, error) {
	repoQuestions, err := app.Repo.ListAll(context.Background())
	if err != nil {
		return nil, err
	}

	var modelQuestions []models.Question
	for _, repoQ := range repoQuestions {
		modelQ, err := toModelQuestion(repoQ)
		if err != nil {
			return nil, err
		}
		modelQuestions = append(modelQuestions, *modelQ)
	}

	// app.Renderer.AsMarkdown()
	// app.Renderer.QuestionsAsTable()
	return modelQuestions, nil
}

func parseFunctionName(snippet string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "src.go", snippet, 0)
	if err != nil {
		return "", fmt.Errorf("failed to parse go snippet: %w", err)
	}

	var functionNames []string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			functionNames = append(functionNames, fn.Name.Name)
		}
		return true
	})

	if len(functionNames) == 0 {
		return "", fmt.Errorf("no functions found in go snippet")
	}

	return functionNames[0], nil
}

// RefreshCookies fetches the session and csrf cookies from the browser and updates the app's config.
func (app *App) RefreshCookies() error {
	cookies, err := browser.GetCookies()
	if err != nil {
		return fmt.Errorf("failed to get browser cookies: %v", err)
	}

	sessionID, hasSession := cookies["LEETCODE_SESSION"]
	csrfToken, hasCSRF := cookies["csrftoken"]

	if !hasSession || !hasCSRF {
		return fmt.Errorf("required leetcode cookies not found in browser")
	}

	app.lcs.SetCookies(sessionID, csrfToken)
	return app.Config.UpdateSession(sessionID, csrfToken)
}

func FormatPathForDisplay(path string) string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}

	return path
}
