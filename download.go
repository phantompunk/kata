package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/spf13/cobra"
)

const PY_EXT = ".py"
const API_URL = "https://leetcode.com/graphql"
const GO_BASE = `package main
%s
`
const GO_TEST_BASE = `package main

func Test%s(t *testing.T) {
        testCases := []struct {
                name   string
                nums   []int
                target int
                expect []int
        }{{}}

        for _, tc := range testCases {
                t.Run(tc.name, func(t *testing.T) {
                        result := %s(tc.nums, tc.target)
                        if !reflect.DeepEqual(result, tc.expect) {
                                t.Errorf("%s(%v, %d) = %v, expected %v", tc.nums, tc.target, result, tc.expect)
                        }
                })
        }
}
`
const PY_TEST_BASE = `import unittest
from %s import Solution


class TestSolution(unittest.TestCase):
  def test_empty_strings(self):
    self.assertTrue(Solution())
`

var commonLangName = map[string]string{
	"go":     "golang",
	"python": "python3",
}

type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type Response struct {
	Data Data `json:"data"`
}

type Data struct {
	Question Question `json:"question"`
}

type Question struct {
	Content      string        `json:"content"`
	QuestionID   string        `json:"questionId"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type LeetCodeClient struct {
	BaseUrl    string
	HttpClient *http.Client
}

func NewLeetCodeClient(baseURL string, client *http.Client) *LeetCodeClient {
	return &LeetCodeClient{BaseUrl: baseURL, HttpClient: client}
}

func (c LeetCodeClient) FetchProblemInfo(problem, lang string) (*Response, error) {
	query := `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
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

	return &response, nil
	//
	// var content string
	// var snippet string
	//
	// if lcRes.Data.Content != "" {
	// 	content = lcRes.Data.Content
	// }
	// for _, v := range lcRes.Data.Question.CodeSnippets {
	// 	if v.LangSlug == commonLangName[language] {
	// 		snippet = v.Code
	// 		break
	// 	}
	// }
	// return snippet, content, nil
}

func DownloadFunc(cmd *cobra.Command, args []string) error {
	ensureConfig()
	name, err := cmd.Flags().GetString("problem")
	if err != nil {
		return err
	}

	language, err := cmd.Flags().GetString("language")
	if err != nil {
		return err
	}

	// fetch code snippet
	snippet, content, err := fetchProblem(name, language)
	if err != nil {
		return err
	}

	fname, err := stubProblem(name, language, snippet, content)
	if err != nil {
		return err
	}
	fmt.Println("Problem stubbed at", fname)
	return nil
}

func fetchProblem(name, language string) (string, string, error) {
	return "", "", nil
}

// 	query := `query questionEditorData($titleSlug: String!) {
//   question(titleSlug: $titleSlug) {
//     questionId
//     codeSnippets {
//       langSlug
//       code
//     }
//   }
// }`
// 	body, err := json.Marshal(Request{Query: query, Variables: map[string]any{
// 		"titleSlug": name,
// 	}})
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(body))
// 	if err != nil {
// 		return "", "", err
// 	}
// 	req.Header.Set("Content-Type", "application/json")
// 	client := &http.Client{}
// 	res, err := client.Do(req)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	defer res.Body.Close()
//
// 	body, err = io.ReadAll(res.Body)
// 	var lcRes Response
// 	err = json.Unmarshal(body, &lcRes)
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	var content string
// 	var snippet string
//
// 	if lcRes.Data.Content != "" {
// 		content = lcRes.Data.Content
// 	}
// 	for _, v := range lcRes.Data.Question.CodeSnippets {
// 		if v.LangSlug == commonLangName[language] {
// 			snippet = v.Code
// 			break
// 		}
// 	}
// 	return snippet, content, nil
// }

func languageExtension(lang string) string {
	extMap := map[string]string{
		"python":  ".py",
		"python3": ".py",
		"go":      ".go",
	}
	return extMap[lang]
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
		testSnippet := fmt.Sprintf(PY_TEST_BASE, name)
		err := os.WriteFile(testFileName, []byte(testSnippet), os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	return fileName, nil
}

var numberToString = map[string]string{"1": "one", "2": "two", "3": "three", "4": "four"}

func convertNumberToWritten(name string) string {
	letters := strings.Split(name, "")
	for i, letter := range letters {
		if hasNumber(letter) {
			written := numberToString[letter]
			letters[i] = written
		}
	}
	return strings.Join(letters, "")
}

func formatProblemName(name string) string {
	if hasNumber(name) {
		return convertNumberToWritten(name)
	}
	return strings.ReplaceAll(name, "-", "_")
}

func hasNumber(name string) bool {
	for _, char := range name {
		if '0' <= char && char <= '9' {
			return true
		}
	}
	return false
}
