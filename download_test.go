package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Fetch a reponse
func TestGetCodeResponse(t *testing.T) {
	mockResponse := `{"data":{"question":{"questionId":"2","content":"Sample Problem","codeSnippets":[{"langSlug":"cpp","code":"Function Sample()"}]}}}`
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(mockResponse)),
					Header:     make(http.Header),
				}
			},
		},
	}

	client := NewLeetCodeClient("", mockClient)
	respo, err := client.FetchProblemInfo("test", "go")
	if err != nil {
		t.Fatal("Failed ", err.Error())
	}
	if respo.Data.Question.Content != "Sample Problem" {
		t.Errorf("got %q, want %q", respo.Data.Question.Content, "Sample problem")
	}
}

// Stubs a problem with readme, solution, test

// Renders file

// Converts numeric name to word: 3sum -> threesum
func TestConvertProblemToPackage(t *testing.T) {
	testCases := []struct {
		problemName string
		packageName string
	}{
		{"4sum", "foursum"},
		{"3sum", "threesum"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("convert %s -> %s", tc.problemName, tc.packageName), func(t *testing.T) {
			got := convertNumberToWritten(tc.problemName)
			want := tc.packageName
			if got != want {
				t.Errorf("got %s want %s", got, want)
			}
		})
	}
}
