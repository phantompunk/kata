package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/phantompunk/kata/internal/app"
	"github.com/spf13/cobra"
)

const API_URL = "https://leetcode.com/graphql"

var commonLangName = map[string]string{
	"go":     "golang",
	"python": "python3",
}

// func NewLeetCodeClient(baseURL string, client *http.Client, fileSystem afero.Fs) *LeetCodeClient {
// 	if baseURL == "" {
// 		baseURL = API_URL
// 	}
//
// 	if _, err := fileSystem.Stat(cfg.Workspace); os.IsNotExist(err) {
// 		fileSystem.MkdirAll(cfg.Workspace, os.ModePerm)
// 	}
//
// 	if client == nil {
// 		client = http.DefaultClient
// 	}
// 	return &LeetCodeClient{BaseUrl: baseURL, HttpClient: client, fileSystem: fileSystem}
// }

// func (c LeetCodeClient) FetchProblemInfo(problem, lang string) (*Problem, error) {
// 	query := `query questionEditorData($titleSlug: String!) {
//   question(titleSlug: $titleSlug) {
//     questionId
//     content
//     titleSlug
//     codeSnippets {
//       langSlug
//       code
//     }
//   }
// }`
//
// 	variables := map[string]any{"titleSlug": problem}
// 	body, err := json.Marshal(Request{Query: query, Variables: variables})
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Set("Content-Type", "application/json")
// 	res, err := c.HttpClient.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer res.Body.Close()
//
// 	body, err = io.ReadAll(res.Body)
// 	var response Response
// 	err = json.Unmarshal(body, &response)
// 	if err != nil || response.Data == nil || response.Data.Question == nil {
// 		return nil, fmt.Errorf("problem not found")
// 	}
//
// 	return response.ToProblem(commonLangName[lang]), nil
// }

func DownloadFunc(cmd *cobra.Command, args []string) error {
	kata, err := app.New()
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("problem")
	if err != nil || name == "" {
		cmd.Usage()
		fmt.Println()
		return fmt.Errorf("Use \"kata download --problem two-sum\" to download and stub a problem")
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil || language == "" {
		language = kata.Config.Language
	}

	open, err := cmd.Flags().GetBool("open")
	if err != nil || !open && kata.Config.OpenInEditor {
		open = kata.Config.OpenInEditor
	}

	// fetch code snippet
	question, err := kata.Questions.FetchQuestion(name, language)
	if err != nil && question == nil {
		return fmt.Errorf("Problem %q not found", name)
	}

	problem := question.ToProblem(language)
	err = kata.StubProblem(problem)
	if err != nil {
		return err
	}

	fmt.Println("Problem stubbed at", filepath.Join(kata.Config.Workspace, problem.SolutionFilePath()))
	if open {
		openWithEditor(filepath.Join(kata.Config.Workspace, problem.SolutionFilePath()))
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
