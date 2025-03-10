package models

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/phantompunk/kata/internal/assert"
)

// mockRoundTripper implements http.RoundTripper for testing.
type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) *http.Response
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req), nil
}

func newMockClient(response string) *http.Client {
	return &http.Client{
		Transport: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader([]byte(response))),
					Header:     make(http.Header),
				}
			},
		},
	}
}

func TestFetchProblem(t *testing.T) {
	title, id, content, langslug, code := "sample", "2", "Sample Problem", "golang", "Function Sample()"
	mockResponse := formatMockResponse(t, title, id, content, langslug, code)
	mockClient := newMockClient(mockResponse)
	mock := QuestionModel{Client: mockClient}
	question, err := mock.FetchQuestion("test", "go")

	assert.NilError(t, err)
	assert.Equal(t, question.ID, id)
	assert.Equal(t, question.Content, content)
	assert.Equal(t, question.CodeSnippets[0].Code, code)
	assert.Equal(t, question.CodeSnippets[0].LangSlug, langslug)
}

func formatMockResponse(t testing.TB, title, id, content, langslug, code string) string {
	t.Helper()
	payload := `{"data":{"question":{"titleSlug":"%s","questionId":"%s","content":"%s","codeSnippets":[{"langSlug":"%s","code":"%s"}]}}}`
	return fmt.Sprintf(payload, title, id, content, langslug, code)
}
