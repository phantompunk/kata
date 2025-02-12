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

	"github.com/spf13/cobra"
)

const PY_EXT = ".py"
const API_URL = "https://leetcode.com/graphql"
const TEST_BASE = `import unittest
from %s import Solution


class TestSolution(unittest.TestCase):
  def test_empty_strings(self):
    self.assertTrue(Solution())
`

type Request struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type Response struct {
	Data struct {
		Question struct {
			QuestionID   string
			CodeSnippets []CodeSnippet
		}
	}
}

func DownloadFunc(cmd *cobra.Command, args []string) error {
	ensureConfig()
	name, err := cmd.Flags().GetString("problem")
	if err != nil {
		return err
	}

	// fetch code snippet
	snippet, err := fetchProblem(name)
	if err != nil {
		return err
	}

	fname, err := stubProblem(name, snippet)
	fmt.Println("Problem stubbed at", fname)
	return nil
}

func fetchProblem(name string) (string, error) {
	query := `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    codeSnippets {
      langSlug
      code
    }
  }
}`
	body, err := json.Marshal(Request{Query: query, Variables: map[string]any{
		"titleSlug": name,
	}})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", API_URL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err = io.ReadAll(res.Body)
	var lcRes Response
	err = json.Unmarshal(body, &lcRes)
	if err != nil {
		return "", err
	}

	var snippet string
	for _, v := range lcRes.Data.Question.CodeSnippets {
		if v.LangSlug == "python3" {
			snippet = v.Code
			break
		}
	}
	return snippet, nil
}

func stubProblem(name, snippet string) (string, error) {
	name = formatProblemName(name)
	fileName := filepath.Join(cfg.Workspace, "python", fmt.Sprintf("%s%s", name, ".py"))
	testFileName := filepath.Join(cfg.Workspace, "python", fmt.Sprintf("%s%s", name, "_test.py"))

	err := os.WriteFile(fileName, []byte(snippet), os.ModePerm)
	if err != nil {
		return "", err
	}

	if _, err = os.Stat(testFileName); err != nil {
		testSnippet := fmt.Sprintf(TEST_BASE, name)
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
