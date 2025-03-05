package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/spf13/afero"
)

var (
	testFS  afero.Fs
	baseDir string = "templates"
)

func init() {
	testFS = afero.NewMemMapFs()
}

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

// Fetch a reponse
func TestGetCodeResponse(t *testing.T) {
	title, id, content, langslug, code := "sample", "2", "Sample Problem", "golang", "Function Sample()"
	mockResponse := formatMockResponse(t, title, id, content, langslug, code)
	mockClient := newMockClient(mockResponse)
	client := NewLeetCodeClient("", mockClient, testFS)
	problem, err := client.FetchProblemInfo("test", "go")

	assertNilError(t, err)
	t.Log(problem.Code)
	assertEqual(t, problem.QuestionID, id)
	assertEqual(t, problem.Content, content)
	assertEqual(t, problem.LangSlug, langslug)
}

// Stubs a problem with readme, solution, test
func TestStubProblem(t *testing.T) {
	lc := NewLeetCodeClient("", nil, testFS)
	problem := Problem{Content: "test", Code: "func sum(a, b int) int {}", QuestionID: "1", LangSlug: "go"}
	err := lc.StubProblem(problem)

	assertNilError(t, err)
	assertExists(t, problem.SolutionFilePath())
	assertExists(t, problem.TestFilePath())
	assertExists(t, problem.ReadmeFilePath())
}

// Renders file
func TestRender(t *testing.T) {
	t.Run("Render go solution stub", func(t *testing.T) {
		problem := Problem{Content: "test", Code: "func sum(a, b int) int {}", QuestionID: "1", LangSlug: "go"}
		want := `package kata
func sum(a, b int) int {}
`
		buf := bytes.Buffer{}
		r := Renderer{}
		err := r.Render(&buf, problem, "solution")

		assertNilError(t, err)
		assertEqual(t, buf.String(), want)
	})

	t.Run("Render python solution stub", func(t *testing.T) {
		problem := Problem{Content: "test", Code: "class Solution(object):\\ndef sum(self, a, b):", QuestionID: "1", LangSlug: "python"}
		want := `class Solution(object):\ndef sum(self, a, b):`
		buf := bytes.Buffer{}
		r := Renderer{}
		err := r.Render(&buf, problem, "solution")

		assertNilError(t, err)
		assertEqual(t, buf.String(), want)
	})

	t.Run("Render go readme stub", func(t *testing.T) {
		problem := Problem{Content: "test", Code: "func sum(a, b int) int {}", QuestionID: "1", LangSlug: "go"}
		want := `test`
		buf := bytes.Buffer{}
		r := Renderer{}
		err := r.Render(&buf, problem, "readme")

		assertNilError(t, err)
		assertEqual(t, buf.String(), want)
	})

	t.Run("Render go test stub", func(t *testing.T) {
		problem := Problem{Content: "test", Code: "func sum(a, b int) int {}", QuestionID: "1", LangSlug: "go"}
		want := `package kata

import (
    "testing"
)

func TestSol(t *testing.T) {
    testCases := []struct {
        nums   []int
        want   []int
    }{
        {[]int{3}, []int{3}},
    }

    for _, tc := range testCases {
        got := Sol(tc.nums, tc.target)
        if !reflect.DeepEqual(got, tc.want) {
            t.Errorf("got %v want %v", got, tc.want)
        }
    }
}
`
		buf := bytes.Buffer{}
		r := Renderer{}
		err := r.Render(&buf, problem, "test")

		assertNilError(t, err)
		assertEqual(t, buf.String(), want)
	})
}

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

func formatMockResponse(t testing.TB, title, id, content, langslug, code string) string {
	t.Helper()
	payload := `{"data":{"question":{"titleSlug":"%s","questionId":"%s","content":"%s","codeSnippets":[{"langSlug":"%s","code":"%s"}]}}}`
	return fmt.Sprintf(payload, title, id, content, langslug, code)
}

func assertEqual(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func assertNilError(t *testing.T, actual error) {
	t.Helper()

	if actual != nil {
		t.Errorf("got %v; expected: nil", actual)
	}
}

func assertExists(t *testing.T, stubFilePath string) {
	t.Helper()

	_, err := testFS.Stat(stubFilePath)
	if err != nil {
		t.Error("sample template not found")
	}

}

func TestProblem(t *testing.T) {
	problem := Problem{TitleSlug: "two-sum", Content: "test", Code: "class Solution(object):\\ndef sum(self, a, b):", QuestionID: "1", LangSlug: "python"}
	got := problem.SolutionFilePath()
	want := "python/two_sum/two_sum.py"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// func TestSaveToFile(t *testing.T) {
// 	filename := "two_sum.go"
// 	content := "func test(){}"
//
// }
