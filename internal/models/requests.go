package models

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

type Request struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

type Response struct {
	Data *Data `json:"data"`
}

type Data struct {
	Question *Question `json:"question"`
}

func GetLangName(language string) string {
	var commonName string
	switch language {
	case "go":
		commonName = "golang"
	case "python":
		commonName = "python3"
	case "c++":
		commonName = "cpp"
	case "c#":
		commonName = "csharp"
	default:
		commonName = language
	}
	return commonName
}

func parseFunctionName(snippets []CodeSnippet) (string, error) {
	var goSnippet string
	for _, snippet := range snippets {
		if snippet.LangSlug == "golang" {
			goSnippet = snippet.Code
			break // Added break, because we only need one golang snippet.
		}
	}

	if goSnippet == "" {
		return "", fmt.Errorf("no golang snippet found")
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "src.go", goSnippet, 0)
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

func (r *Response) GetQuestion(language string) (*Question, error) {
	if r != nil && r.Data != nil && r.Data.Question != nil {
		var selected CodeSnippet

		for _, snippet := range r.Data.Question.CodeSnippets {
			if snippet.LangSlug == GetLangName(language) {
				selected = snippet
				r.Data.Question.CodeSnippets = []CodeSnippet{selected}
				return r.Data.Question, nil
			}
		}
	}
	return nil, fmt.Errorf("Code snippet for %q not found", language)
}
