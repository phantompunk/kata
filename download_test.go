package main

// var testFS afero.Fs
// var r Renderer
// var problem = Problem{TitleSlug: "two-sum", Content: "test", Code: "func sum(a, b int) int {}", QuestionID: "1", LangSlug: "go"}
// var pythonProblem = Problem{TitleSlug: "two-sum", Content: "test", Code: "class Solution(object):\n\tdef sum(self, a, b):\n", QuestionID: "1", LangSlug: "python"}
//
// func init() {
// 	testFS = afero.NewMemMapFs()
// }
//
// // mockRoundTripper implements http.RoundTripper for testing.
// type mockRoundTripper struct {
// 	roundTripFunc func(req *http.Request) *http.Response
// }
//
// func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
// 	return m.roundTripFunc(req), nil
// }
//
// func newMockClient(response string) *http.Client {
// 	return &http.Client{
// 		Transport: &mockRoundTripper{
// 			roundTripFunc: func(req *http.Request) *http.Response {
// 				return &http.Response{
// 					StatusCode: http.StatusOK,
// 					Body:       io.NopCloser(bytes.NewReader([]byte(response))),
// 					Header:     make(http.Header),
// 				}
// 			},
// 		},
// 	}
// }
//
// // Fetch a reponse
// func TestGetCodeResponse(t *testing.T) {
// 	title, id, content, langslug, code := "sample", "2", "Sample Problem", "golang", "Function Sample()"
// 	mockResponse := formatMockResponse(t, title, id, content, langslug, code)
// 	mockClient := newMockClient(mockResponse)
// 	client := NewLeetCodeClient("", mockClient, testFS)
// 	problem, err := client.FetchProblemInfo("test", "go")
//
// 	assertNilError(t, err)
// 	assertEqual(t, problem.QuestionID, id)
// 	assertEqual(t, problem.Content, content)
// 	assertEqual(t, problem.LangSlug, langslug)
// }
//
// // Stubs a problem with readme, solution, test
// func TestStubProblem(t *testing.T) {
// 	lc := NewLeetCodeClient("", nil, testFS)
// 	err := lc.StubProblem(problem)
//
// 	assertNilError(t, err)
// 	assertExists(t, problem.SolutionFilePath())
// 	assertExists(t, problem.TestFilePath())
// 	assertExists(t, problem.ReadmeFilePath())
// }
//
// // Renders file
// func TestRender(t *testing.T) {
// 	t.Run("Render go solution stub", func(t *testing.T) {
// 		buf := bytes.Buffer{}
// 		err := r.Render(&buf, problem, "solution")
//
// 		assertNilError(t, err)
// 		approvals.VerifyString(t, buf.String())
// 	})
//
// 	t.Run("Render python solution stub", func(t *testing.T) {
// 		buf := bytes.Buffer{}
// 		err := r.Render(&buf, pythonProblem, "solution")
//
// 		assertNilError(t, err)
// 		approvals.VerifyString(t, buf.String())
// 	})
//
// 	t.Run("Render go readme stub", func(t *testing.T) {
// 		want := `test`
// 		buf := bytes.Buffer{}
// 		err := r.Render(&buf, problem, "readme")
//
// 		assertNilError(t, err)
// 		assertEqual(t, buf.String(), want)
// 	})
//
// 	t.Run("Render go test stub", func(t *testing.T) {
// 		buf := bytes.Buffer{}
// 		err := r.Render(&buf, problem, "test")
//
// 		assertNilError(t, err)
// 		// approvals.VerifyString(t, buf.String())
// 	})
// }
//
// // Converts numeric name to word: 3sum -> threesum
// func TestConvertProblemToPackage(t *testing.T) {
// 	testCases := []struct {
// 		problemName string
// 		packageName string
// 	}{
// 		{"4sum", "foursum"},
// 		{"3sum", "threesum"},
// 	}
//
// 	for _, tc := range testCases {
// 		t.Run(fmt.Sprintf("convert %s -> %s", tc.problemName, tc.packageName), func(t *testing.T) {
// 			got := convertNumberToWritten(tc.problemName)
// 			want := tc.packageName
// 			assertEqual(t, got, want)
// 		})
// 	}
// }
//
// func formatMockResponse(t testing.TB, title, id, content, langslug, code string) string {
// 	t.Helper()
// 	payload := `{"data":{"question":{"titleSlug":"%s","questionId":"%s","content":"%s","codeSnippets":[{"langSlug":"%s","code":"%s"}]}}}`
// 	return fmt.Sprintf(payload, title, id, content, langslug, code)
// }
//
// func assertEqual(t *testing.T, got, want string) {
// 	t.Helper()
// 	if got != want {
// 		t.Errorf("got %q, want %q", got, want)
// 	}
// }
//
// func assertNilError(t *testing.T, actual error) {
// 	t.Helper()
//
// 	if actual != nil {
// 		t.Errorf("got %v; expected: nil", actual)
// 	}
// }
//
// func assertExists(t *testing.T, stubFilePath string) {
// 	t.Helper()
//
// 	_, err := testFS.Stat(stubFilePath)
// 	if err != nil {
// 		t.Error("sample template not found")
// 	}
//
// }
//
// func TestProblem(t *testing.T) {
// 	got := pythonProblem.SolutionFilePath()
// 	want := "two_sum/two_sum.py"
// 	assertEqual(t, got, want)
// }
