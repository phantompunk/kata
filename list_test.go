package main

import "testing"

var goProblem = Question{QuestionID: "1", Content: "demo", TitleSlug: "demo", Title: "Demo", CodeSnippets: []CodeSnippet{{Code: "func sample()", LangSlug: "go"}}}
var pyProblem = Question{QuestionID: "2", Content: "sample", TitleSlug: "sample", Title: "Sample", CodeSnippets: []CodeSnippet{{Code: "def sample()", LangSlug: "python"}}}

func TestList(t *testing.T) {
	// given a list of problems
	// buf := bytes.Buffer{}
	_ = []Question{goProblem, pyProblem}

	// pretty print as a table
	// got := printAsTable(buf, given)
	// assertEqual(t, buf.String(), "test")
}
