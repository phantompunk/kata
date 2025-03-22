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
	testDB := newTestDB(t)
	t.Run("Fetch a new problem", func(t *testing.T) {
		title, id, diff, content, langslug, code := "5sum", "123", "Easy", "Sample Problem", "golang", "func fiveSum() {}"
		mockClient := newMockClient(mockResponse(t, title, id, diff, content, langslug, code))
		mock := QuestionModel{Client: mockClient, DB: testDB}
		question, err := mock.FetchQuestion("5sum")

		assert.NilError(t, err)
		assert.Equal(t, question.ID, id)
		assert.Equal(t, question.Content, content)
		assert.Equal(t, question.CodeSnippets[0].Code, code)
		assert.Equal(t, question.CodeSnippets[0].LangSlug, langslug)
	})

	t.Run("Fetch an existing problem", func(t *testing.T) {
		var mockClient *http.Client
		mock := QuestionModel{Client: mockClient, DB: testDB}
		question, err := mock.FetchQuestion("two-sum")

		assert.NilError(t, err)
		assert.Equal(t, question.ID, "36")
		assert.Equal(t, question.Content, "Sample problem description")
		assert.Equal(t, question.CodeSnippets[0].Code, "Function Sample()")
		assert.Equal(t, question.CodeSnippets[0].LangSlug, "cpp")
	})

	t.Run("Fetch previous problem", func(t *testing.T) {
		title, id, diff, content, langslug, code := "5sum", "123", "Easy", "Sample Problem", "golang", "func fiveSum() {}"
		mockClient := newMockClient(mockResponse(t, title, id, diff, content, langslug, code))
		mock := QuestionModel{Client: mockClient, DB: testDB}
		_, err := mock.FetchQuestion("5sum")

		var fakeClient *http.Client
		mock.Client = fakeClient
		prev, err := mock.Get("5sum")
		assert.NilError(t, err)
		assert.Equal(t, prev.ID, id)
		assert.Equal(t, prev.Content, content)
		assert.Equal(t, prev.CodeSnippets[0].Code, code)
		assert.Equal(t, prev.CodeSnippets[0].LangSlug, langslug)
	})
}

func mockResponse(t testing.TB, title, id, difficulty, content, langslug, code string) string {
	t.Helper()
	payload := `{"data":{"question":{"titleSlug":"%s","questionId":"%s","content":"%s","difficulty":"%s","codeSnippets":[{"langSlug":"%s","code":"%s"}]}}}`
	return fmt.Sprintf(payload, title, id, content, difficulty, langslug, code)
}

func TestQuestionModelExists(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want bool
	}{
		{
			name: "Valid Slug",
			slug: "two-sum",
			want: true,
		},
		{
			name: "invalid Slug ",
			slug: "three-sum",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t)
			m := QuestionModel{DB: db}

			exists, err := m.Exists(tc.slug)
			assert.Equal(t, exists, tc.want)
			assert.NilError(t, err)
		})
	}
}

func TestQuestionModelGet(t *testing.T) {
	tests := []struct {
		name string
		slug string
		want string
	}{
		{
			name: "Valid Slug",
			slug: "two-sum",
			want: "Two Sum",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t)
			m := QuestionModel{DB: db}

			question, err := m.Get(tc.slug)
			assert.Equal(t, question.Title, tc.want)
			assert.NilError(t, err)
		})
	}
}

func TestQuestionModelGetWithStatus(t *testing.T) {
	tests := []struct {
		name   string
		slug   string
		want   string
		solved string
	}{
		{
			name:   "Valid Slug",
			slug:   "two-sum",
			want:   "Two Sum",
			solved: "java",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db := newTestDB(t)
			m := QuestionModel{DB: db}

			questions, err := m.GetAllWithStatus([]string{"go", "python", "java"})

			assert.Equal(t, len(questions), 1)
			assert.Equal(t, questions[0].Title, tc.want)
			assert.Equal(t, questions[0].LangStatus[tc.solved], true)
			assert.NilError(t, err)
		})
	}
}
