package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const API_URL = "https://leetcode.com/graphql"

var commonLangName = map[string]string{
	"go":     "golang",
	"python": "python3",
}

func NewLeetCodeClient(baseURL string, client *http.Client, fileSystem afero.Fs) *LeetCodeClient {
	if baseURL == "" {
		baseURL = API_URL
	}

	if _, err := fileSystem.Stat(cfg.Workspace); os.IsNotExist(err) {
		fileSystem.MkdirAll(cfg.Workspace, os.ModePerm)
	}

	if client == nil {
		client = http.DefaultClient
	}
	return &LeetCodeClient{BaseUrl: baseURL, HttpClient: client, fileSystem: fileSystem}
}

func (c LeetCodeClient) FetchProblemInfo(problem, lang string) (*Problem, error) {
	query := `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    content
    titleSlug
    codeSnippets {
      langSlug
      code
    }
  }
}`

	variables := map[string]any{"titleSlug": problem}
	body, err := json.Marshal(Request{Query: query, Variables: variables})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.ToProblem(commonLangName[lang]), nil
}

func DownloadFunc(cmd *cobra.Command, args []string) error {
	ensureConfig()

	fileSystem := afero.NewOsFs()

	leet := NewLeetCodeClient(API_URL, nil, fileSystem)

	name, err := cmd.Flags().GetString("problem")
	if err != nil {
		return err
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil {
		return err
	}

	// fetch code snippet
	problem, err := leet.FetchProblemInfo(name, language)
	if err != nil {
		return err
	}

	err = leet.StubProblem(*problem)
	if err != nil {
		return err
	}

	fmt.Println("Problem stubbed at", problem.SolutionFilePath())
	return nil
}

func languageExtension(lang string) string {
	extMap := map[string]string{
		"python":  ".py",
		"python3": ".py",
		"go":      ".go",
	}
	return extMap[lang]
}

type Filer struct {
	path  string
	file  afero.File
	ttype string
}

func (c *LeetCodeClient) StubProblem(problem Problem) error {
	// fileMap := map[string]Filer{
	// 	"solution": Filer{solutionPath,file,},
	// }
	if err := c.fileSystem.MkdirAll(problem.DirFilePath(), os.ModePerm); err != nil {
		fmt.Println("Error making dirs")
		return err
	}

	file, err := c.fileSystem.Create(problem.SolutionFilePath())
	if err != nil {
		fmt.Println("Error making file at", problem.SolutionFilePath())
		return err
	}
	test, err := c.fileSystem.Create(problem.TestFilePath())
	if err != nil {
		fmt.Println("Error making test dirs")
		return err
	}
	readme, err := c.fileSystem.Create(problem.ReadmeFilePath())
	if err != nil {
		fmt.Println("Error making readme dirs")
		return err
	}

	r := Renderer{}
	r.Render(file, problem, "solution")
	r.Render(test, problem, "test")
	r.Render(readme, problem, "readme")
	if r.error != nil {
		return r.error
	}

	return nil
}

func stubProblem(name, language, snippet, content string) (string, error) {
	name = formatProblemName(name)
	ext := languageExtension(language)
	fileName := filepath.Join(cfg.Workspace, language, fmt.Sprintf("%s%s", name, ext))
	readMe := filepath.Join(cfg.Workspace, language, "README.md")
	testFileName := filepath.Join(cfg.Workspace, language, fmt.Sprintf("%s_test%s", name, ext))

	fDir := filepath.Dir(fileName)
	if dirErr := os.MkdirAll(fDir, os.ModePerm); dirErr != nil {
		return "", fmt.Errorf("Could not create config directory %v", dirErr)
	}

	err := os.WriteFile(fileName, []byte(snippet), os.ModePerm)
	if err != nil {
		return "", err
	}

	markdown, err := htmltomarkdown.ConvertString(content)
	err = os.WriteFile(readMe, []byte(markdown), os.ModePerm)
	if err != nil {
		return "", err
	}

	if _, err = os.Stat(testFileName); err != nil {
		testSnippet := fmt.Sprint(name)
		err := os.WriteFile(testFileName, []byte(testSnippet), os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return fileName, nil
}

var (
	//go:embed "templates/*"
	postTemplates embed.FS
)

type Renderer struct {
	error error
}

func langTemplates(lang string) (string, string) {
	var solBlock string
	var testBlock string
	switch lang {
	case "go", "golang":
		solBlock = "golang"
		testBlock = "gotest"
	case "python", "python3":
		solBlock = "python"
		testBlock = "pytest"
	default:
		solBlock = lang
		testBlock = lang
	}
	return solBlock, testBlock
}

func (r *Renderer) Render(w io.Writer, problem Problem, templateType string) error {
	if r.error != nil {
		return r.error
	}
	templ, err := template.New(templateType).ParseFS(postTemplates, "templates/*.gohtml")
	if err != nil {
		return err
	}

	var langBlock string
	if templateType == "solution" || templateType == "test" {
		sol, test := langTemplates(problem.LangSlug)
		if templateType == "solution" {
			langBlock = sol
		} else {
			langBlock = test
		}
	}

	if langBlock != "" {
		var buf bytes.Buffer
		err = templ.ExecuteTemplate(&buf, langBlock, problem)
		if err != nil {
			return err
		}
		problem.Code = buf.String()
	}

	if templateType == "readme" {
		markdown, err := htmltomarkdown.ConvertString(problem.Content)
		if err != nil {
			return err
		}

		problem.Content = markdown
	}

	if err = templ.ExecuteTemplate(w, templateType, problem); err != nil {
		return err
	}
	return nil
}
