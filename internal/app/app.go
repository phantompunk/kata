package app

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/db"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/renderer"
	"github.com/phantompunk/kata/internal/repository"
	"github.com/spf13/afero"
)

const (
	LoginUrl = "https://leetcode.com/accounts/login/"
	// Maximum number of attempts to check test status
	MaxTestAttempts = 10
	// Interval between test status checks
	TestPollInterval = 500 * time.Millisecond
)

type AppOptions struct {
	Problem  string
	Language string
	Open     bool
	Force    bool
}

type App struct {
	Config   *config.Config
	repo     *repository.Queries
	lcs      *leetcode.Service
	Renderer *renderer.Renderer
	fs       afero.Fs
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

	renderer, err := renderer.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create renderer: %w", err)
	}

	return &App{
		Config:   &cfg,
		repo:     repo,
		lcs:      lcs,
		Renderer: renderer,
		fs:       afero.NewOsFs(),
	}, nil
}

func (app *App) DownloadQuestion(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	if !opts.Open && app.Config.OpenInEditor {
		opts.Open = true
	}

	question, err := app.GetQuestion(opts.Problem, opts.Language, opts.Force)
	if err != nil {
		return fmt.Errorf("fetching question %q: %w", opts.Problem, err)
	}

	if err := app.Stub(question, opts); err != nil {
		return fmt.Errorf("stubbing problem %q: %w", opts.Problem, err)
	}

	if opts.Open {
		// :TODO: Open the solution file in the editor
	}

	return nil
}

func (app *App) ListQuestions() error {
	questions, err := app.repo.GetAllWithStatus(context.Background(), app.Config.Tracks)
	if err != nil {
		return fmt.Errorf("listing questions: %w", err)
	}

	if err := app.Renderer.QuestionsAsTable(questions, app.Config.Tracks); err != nil {
		return fmt.Errorf("rendering questions as table: %w", err)
	}

	return nil
}

func (app *App) Login(opts AppOptions) error {
	if app.Config.IsSessionValid() && !opts.Force {
		fmt.Println("You are already logged in")
		return nil
	}

	// TODO: Improve user error message
	if err := app.RefreshCookies(); err != nil {
		return fmt.Errorf("could not authenticate using browser cookies: %w\nPlease login manually at %s", err, LoginUrl)
	}

	valid, err := app.CheckSession()
	if err != nil {
		return fmt.Errorf("failed to check session: %w\nPlease login manually at %s", err, LoginUrl)
	}

	if !valid {
		app.ClearCookies()
		return fmt.Errorf("session cookies are invalid.\nPlease login manually at %s", LoginUrl)
	}

	fmt.Println("Successfully logged in using browser cookies.")
	return nil
}

func (app *App) Quiz(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	question, err := app.repo.GetRandom(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get random question: %w", err)
	}

	problem := question.ToProblem(app.Config.Workspace, opts.Language)
	if app.Config.OpenInEditor || opts.Open {
		if err := openWithEditor(problem.SolutionPath); err != nil {
			return fmt.Errorf("failed to open solution file in editor: %w", err)
		}
	}
	return nil
}

func openWithEditor(pathToFile string) error {
	textEditor := findTextEditor()

	command := exec.Command(textEditor, pathToFile)
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	err := os.Chdir(filepath.Dir(pathToFile))
	err = command.Run()
	if err != nil {
		return err
	}
	return nil
}

func findTextEditor() string {
	if isCommandAvailable("nvim") {
		return "nvim"
	} else if isCommandAvailable("vim") {
		return "vim"
	} else if isCommandAvailable("nano") {
		return "nano"
	} else if isCommandAvailable("editor") {
		return "editor"
	} else {
		return "vi"
	}
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("command", "-v", name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (app *App) Test(opts AppOptions) error {
	if opts.Language == "" {
		opts.Language = app.Config.Language
	}

	if !app.Config.IsSessionValid() {
		return fmt.Errorf("session is not valid. Please login using 'kata login' command")
	}

	valid, err := app.CheckSession()
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	if !valid {
		return fmt.Errorf("session is not valid. Please login using 'kata login' command")
	}

	status, err := app.TestSolution(opts.Problem, opts.Language)
	if err != nil {
		return fmt.Errorf("testing solution for %q: %w", opts.Problem, err)
	}

	fmt.Println("Test status:", status)
	return nil
}

// CheckSession checks if the session is valid by pinging the leetcode service
func (app *App) CheckSession() (bool, error) {
	app.lcs.SetCookies(app.Config.SessionToken, app.Config.CsrfToken)
	isValid, err := app.lcs.Ping()
	if err != nil {
		return false, fmt.Errorf("failed to ping leetcode service: %w", err)
	}
	return isValid, nil
}

func (app *App) ClearCookies() error {
	app.Config.SessionToken = ""
	app.Config.CsrfToken = ""
	app.Config.SessionExpires = time.Time{}
	fmt.Println("Cleared cookies from config")
	return app.Config.Update()
}

// :TODO: Move this to a separate package
func toModelQuestion(repoQuestion repository.Question) (*models.Question, error) {
	var modelQuestion models.Question
	modelQuestion.ID = fmt.Sprintf("%d", repoQuestion.Questionid)
	modelQuestion.Title = repoQuestion.Title
	modelQuestion.TitleSlug = repoQuestion.Titleslug
	modelQuestion.Difficulty = repoQuestion.Difficulty
	modelQuestion.FunctionName = repoQuestion.Functionname
	modelQuestion.Content = repoQuestion.Content

	if err := json.Unmarshal([]byte(repoQuestion.Codesnippets), &modelQuestion.CodeSnippets); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(repoQuestion.Testcases), &modelQuestion.TestCases); err != nil {
		return nil, err
	}
	return &modelQuestion, nil
}

func ToProblem(repoQuestion repository.Question, workspace, language string) *models.Problem {
	var problem models.Problem
	problem.QuestionID = fmt.Sprintf("%d", repoQuestion.Questionid)
	problem.TitleSlug = repoQuestion.Titleslug
	problem.FunctionName = repoQuestion.Functionname
	problem.Content = repoQuestion.Content

	var codeSnippets []models.CodeSnippet
	if err := json.Unmarshal([]byte(repoQuestion.Codesnippets), &codeSnippets); err != nil {
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
	params.Questionid = qId
	params.Title = modelQuestion.Title
	params.Titleslug = modelQuestion.TitleSlug
	params.Difficulty = modelQuestion.Difficulty
	params.Functionname = modelQuestion.FunctionName
	params.Content = modelQuestion.Content

	codeSnippets, err := json.Marshal(modelQuestion.CodeSnippets)
	if err != nil {
		return params, err
	}
	params.Codesnippets = string(codeSnippets)

	testCases, err := json.Marshal(modelQuestion.TestCases)
	if err != nil {
		return params, err
	}
	params.Testcases = string(testCases)

	return params, nil
}

func (app *App) GetQuestion(name, language string, force bool) (*models.Question, error) {
	exists, err := app.repo.Exists(context.Background(), name)
	if err != nil {
		return nil, fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 1 && !force {
		repoQuestion, err := app.repo.GetBySlug(context.Background(), name)
		if err != nil {
			return nil, fmt.Errorf("failed to get question details: %w", err)
		}

		modelQuestion, err := toModelQuestion(repoQuestion)
		if err != nil {
			return nil, fmt.Errorf("failed to convert question: %w", err)
		}

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

	app.repo.Create(context.Background(), params)
	return question, nil
}

func (app *App) Stub(question *models.Question, opts AppOptions) error {
	problem := question.ToProblem(app.Config.Workspace, opts.Language)
	if err := app.fs.MkdirAll(problem.DirPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed creating problem directory: %w", err)
	}

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

	if err := app.Renderer.Render(file, problem, "solution"); err != nil {
		return fmt.Errorf("failed to render solution file: %w", err)
	}
	fmt.Println("Problem stubbed at", problem.SolutionPath)

	if err := app.Renderer.Render(test, problem, "test"); err != nil {
		return fmt.Errorf("failed to render test file: %w", err)
	}

	if err := app.Renderer.Render(readme, problem, "readme"); err != nil {
		return fmt.Errorf("failed to render readme file: %w", err)
	}

	return nil
}

// TestSolution tests the solution for a given problem name and language.
func (app *App) TestSolution(name, language string) (string, error) {
	exists, err := app.repo.Exists(context.Background(), name)
	if err != nil {
		return "", fmt.Errorf("failed to check question existence: %w", err)
	}

	if exists == 0 {
		return "", fmt.Errorf("question %q not found", name)
	}

	repoQuestion, err := app.repo.GetBySlug(context.Background(), name)
	if err != nil {
		return "", fmt.Errorf("failed to get question details: %w", err)
	}

	problem := repoQuestion.ToProblem(app.Config.Workspace, language)
	snippet := app.extractSnippet(problem.SolutionPath)

	testStatusUrl, err := app.lcs.Test(problem, language, snippet)
	if err != nil {
		return "", fmt.Errorf("failed to submit solution for testing: %w", err)
	}

	if testStatusUrl == "" {
		return "", fmt.Errorf("empty testStatusUrl received from server")
	}

	res, err := app.lcs.PollTestStatus(testStatusUrl)
	if err != nil {
		return "", fmt.Errorf("failed to poll test status: %w", err)
	}

	if res.Correct {
		return "Passed", nil
	}

	return fmt.Sprintf("Failed: %s. Details: %s", res.StatusMsg, res.TestCase), nil
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

	app.Renderer.Render(&buf, problem, "solution")
	name, err := parseFunctionName(buf.String())
	if err != nil {
		fmt.Println("failed %w", err)
		return ""
	}
	return name
}

func (app *App) PrintQuestionStatus() ([]models.Question, error) {
	repoQuestions, err := app.repo.ListAll(context.Background())
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
	var sessionCookie *kooky.Cookie
	var csrfCookie *kooky.Cookie

	cookiesSeq := kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`)).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "LEETCODE_SESSION" {
			sessionCookie = cookie
			continue
		}

		if cookie.Name == "csrftoken" {
			csrfCookie = cookie
			continue
		}

		if sessionCookie != nil && csrfCookie != nil {
			break
		}
	}

	if sessionCookie == nil || sessionCookie.Expires.Before(time.Now()) {
		return fmt.Errorf("LEETCODE_SESSION missing or expired")
	}

	if csrfCookie == nil {
		return fmt.Errorf("csrftoken missing or expired")
	}

	app.lcs.SetCookies(sessionCookie.Value, csrfCookie.Value)
	return app.Config.UpdateSession(sessionCookie.Value, csrfCookie.Value, sessionCookie.Expires)
}
