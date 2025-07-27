package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/browserutils/kooky"
	"github.com/phantompunk/kata/internal/config"
	"github.com/phantompunk/kata/internal/database"
	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/models"
	"github.com/phantompunk/kata/internal/renderer"
	"github.com/spf13/afero"
)

const LOGIN_URL = "https://leetcode.com/accounts/login/"

type App struct {
	Config    *config.Config
	Questions models.QuestionModel
	lcs       *leetcode.Service
	Renderer  renderer.Renderer
	fs        afero.Fs
}

func New() (*App, error) {
	cfg, err := config.EnsureConfig()
	if err != nil {
		fmt.Println("Failed cfg")
		return nil, err
	}

	db, err := database.EnsureDB(database.GetDbPath())
	if err != nil {
		fmt.Println("Failed db")
		return nil, err
	}

	return &App{
		Config:    &cfg,
		Questions: models.QuestionModel{DB: db, Client: http.DefaultClient},
		lcs:       leetcode.New(leetcode.WithCookies(cfg.SessionToken, cfg.CsrfToken)),
		Renderer:  renderer.New(),
		fs:        afero.NewOsFs(),
	}, nil
}

// CheckSession checks if the session is valid by pinging the leetcode service
func (app *App) CheckSession() bool {
	app.lcs.SetCookies(app.Config.SessionToken, app.Config.CsrfToken)
	return app.lcs.Ping()
}

func (app *App) FetchQuestion(name, language string) (*models.Question, error) {
	// check if question has been saved before
	exists, err := app.Questions.Exists(name)
	if err != nil {
		return nil, err
	}

	if exists {
		return app.Questions.Get(name)
	}

	// fetch the question from leetcode
	question, err := app.lcs.Fetch(name)
	if err != nil {
		return nil, err
	}

	functionName := app.GetFunctionName(question.ToProblem(app.Config.Workspace, "golang"))
	question.FunctionName = functionName

	// save question to database
	_, err = app.Questions.Insert(question)
	if err != nil {
		return nil, err
	}

	return question, nil
}

func (app *App) StubProblem(problem *models.Problem) error {
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

	app.Renderer.Render(file, problem, "solution")
	app.Renderer.Render(test, problem, "test")
	app.Renderer.Render(readme, problem, "readme")
	if app.Renderer.Error != nil {
		return fmt.Errorf("failed to render file %w", app.Renderer.Error)
	}

	return nil
}

// TestSolution tests the solution for a given problem name and language.
func (app *App) TestSolution(name, language string) (string, error) {
	exists, err := app.Questions.Exists(name)
	if err != nil {
		return "", fmt.Errorf("checking existence: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("question %q not found", name)
	}

	question, err := app.Questions.Get(name)
	if err != nil {
		return "", err
	}

	filePath := question.ToProblem(app.Config.Workspace, language).SolutionPath

	snippet := app.extractSnippet(filePath)
	testStatusUrl, err := app.lcs.Test(question, language, snippet)
	if err != nil {
		return "", err
	}

	if testStatusUrl == "" {
		return "", errors.New("empty testStatusUrl received from server")
	}

	res := &models.TestResponse{}
	for range 10 {
		res, err = app.lcs.CheckTestStatus(testStatusUrl)
		if err != nil {
			return "", fmt.Errorf("checking test status: %w", err)
		}
		if res.State == "STARTED" {
			fmt.Print("started")
		}
		if res.State == "PENDING" {
			fmt.Print(".")
		}
		if res.State == "SUCCESS" {
			// fmt.Print("done")
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Print("\n")

	if res.Correct {
		return "Passed", nil
	}
	return fmt.Sprintf("Failed: %s", res.TestCase), nil
}

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
	app.Questions.GetAllWithStatus(app.Config.Tracks)

	// app.Renderer.AsMarkdown()
	// app.Renderer.QuestionsAsTable()
	return []models.Question{}, nil
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

	cookiesSeq := kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(".leetcode.com"), kooky.Name("LEETCODE_SESSION")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "LEETCODE_SESSION" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil || sessionCookie.Expires.Before(time.Now()) {
		return fmt.Errorf("LEETCODE_SESSION missing or expired")
	}

	cookiesSeq = kooky.TraverseCookies(context.TODO(), kooky.Valid, kooky.DomainHasSuffix(`leetcode.com`), kooky.Name("csrftoken")).OnlyCookies()
	for cookie := range cookiesSeq {
		if cookie.Name == "csrftoken" {
			csrfCookie = cookie
			break
		}
	}
	if csrfCookie == nil {
		return fmt.Errorf("csrftoken missing or expired")
	}

	app.lcs.SetCookies(sessionCookie.Value, csrfCookie.Value)
	return app.Config.UpdateSession(sessionCookie.Value, csrfCookie.Value, sessionCookie.Expires)
}
