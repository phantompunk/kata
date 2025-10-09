package leetcode

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const SolutionTask = "judger.judgetask.Judge"

type Question struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	TestCaseList []string      `json:"exampleTestcaseList"`
	Metadata     QuestionMeta  `json:"metadata"`
	LangStatus   map[string]bool
	CreatedAt    string
}

// InternalQuestion is used for unmarshaling the raw metadata field
type InternalQuestion struct {
	ID           string        `json:"questionId"`
	Title        string        `json:"title"`
	TitleSlug    string        `json:"titleSlug"`
	Difficulty   string        `json:"difficulty"`
	Content      string        `json:"content"`
	CodeSnippets []CodeSnippet `json:"codeSnippets"`
	TestCaseList []string      `json:"exampleTestcaseList"`
	RawMetadata  string        `json:"metadata"`
}

type CodeSnippet struct {
	Code     string `json:"code"`
	LangSlug string `json:"langSlug"`
}

type QuestionMeta struct {
	Name string `json:"name"`
}

func (q *Question) UnmarshalJSON(data []byte) error {
	var tmp InternalQuestion
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	if tmp.ID == "" {
		*q = Question{}
		return ErrQuestionNotFound
	}

	if tmp.RawMetadata == "" {
		*q = Question{}
		return ErrMetadataMissing
	}

	q.ID = tmp.ID
	q.Title = tmp.Title
	q.TitleSlug = tmp.TitleSlug
	q.Difficulty = tmp.Difficulty
	q.Content = tmp.Content
	q.CodeSnippets = tmp.CodeSnippets
	q.TestCaseList = tmp.TestCaseList

	if err := json.Unmarshal([]byte(tmp.RawMetadata), &q.Metadata); err != nil {
		return err
	}

	return nil
}

type QuestionReponse struct {
	Data struct {
		Question Question `json:"question"`
	} `json:"data"`
}

type SubmissionID string

// Custom unmarshal to handle both int and string
func (s *SubmissionID) UnmarshalJSON(data []byte) error {
	// If it's quoted, it's a string
	if len(data) > 0 && data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = SubmissionID(str)
		return nil
	}

	// Otherwise, try integer
	var num int64
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*s = SubmissionID(strconv.FormatInt(num, 10))
	return nil
}

type SubmitResponse struct {
	InterpretID  string `json:"interpret_id"`
	SubmissionID int64  `json:"submission_id"`
	Testcase     string `json:"test_case"`
}

func (r SubmitResponse) GetSubmissionID() string {
	return fmt.Sprintf("%d", r.SubmissionID)
}

//		{
//		  "status_code": 15,
//		  "lang": "javascript",
//		  "run_success": false,
//		  "runtime_error": "Line 1: SyntaxError: Invalid or unexpected token",
//		  "full_runtime_error": "Line 1 in solution.js\nvar romanToInt = function(s) {\\n    const values = {\\n        I: 1,\\n        V: 5,\\n        X: 10,\\n        L: 50,\\n        C: 100,\\n        D: 500,\\n        M: 1000\\n    };\\n\\n    let total = 0;\\n\\n    for (let i = 0; i < s.length; i++) {\\n        const current = values[s[i]];\\n        const next = values[s[i + 1]];\\n\\n        // If current < next, it's a subtractive pair (like IV, IX, etc.)\\n        if (next > current) {\\n            total -= current;\\n        } else {\\n            total += current;\\n        }\\n    }\\n\\n    return total;\\n};\n                              ^\nSyntaxError: Invalid or unexpected token\n    at wrapSafe (node:internal/modules/cjs/loader:1486:18)\n    at Module._compile (node:internal/modules/cjs/loader:1528:20)\n    at Object..js (node:internal/modules/cjs/loader:1706:10)\n    at Module.load (node:internal/modules/cjs/loader:1289:32)\n    at Function._load (node:internal/modules/cjs/loader:1108:12)\n    at TracingChannel.traceSync (node:diagnostics_channel:322:14)\n    at wrapModuleLoad (node:internal/modules/cjs/loader:220:24)\n    at Function.executeUserEntryPoint [as runMain] (node:internal/modules/run_main:170:5)\n    at node:internal/main/run_main_module:36:49\nNode.js v22.14.0",
//		  "status_runtime": "N/A",
//		  "correct_answer": false,
//		  "total_correct": 0,
//		  "total_testcases": 3,
//		 "memory_percentile": null,
//		  "pretty_lang": "JavaScript",
//		  "submission_id": "runcode_1759702712.9142835_d55B5qU8tt",
//		  "status_msg": "Runtime Error",
//		  "state": "SUCCESS"
//	   "compile_error": "Line 6: Char 17: undefined: val (solution.go)",
//	   "full_compile_error": "Line 6: Char 17: undefined: val (solution.go)",
//		}
type SubmissionResponse struct {
	SubmissionID       string   `json:"submission_id"`
	QuestionID         string   `json:"question_id"`
	State              string   `json:"state"`      // SUCCESS, FAILED, PENDING
	StatusMsg          string   `json:"status_msg"` // Accepted, Runtime Error, Compile Error
	Correct            bool     `json:"correct_answer"`
	RuntimePercentile  float64  `json:"runtime_percentile"`
	StatusRuntime      string   `json:"status_runtime"`
	MemoryPercentile   float64  `json:"memory_percentile"`
	StatusMemory       string   `json:"status_memory"`
	TotalCorrect       int      `json:"total_correct"`
	TotalTestcases     int      `json:"total_testcases"`
	RuntimeError       string   `json:"runtime_error"`
	FullRuntimeError   string   `json:"full_runtime_error"`
	CompileError       string   `json:"compile_error"`
	FullCompileError   string   `json:"full_compile_error"`
	CodeAnswer         []string `json:"code_answer"`
	ExpectedCodeAnswer []string `json:"expected_code_answer"`
	TaskName           string   `json:"task_name"`
}

func (s SubmissionResponse) ToResult() *SubmissionResult {
	errorMessage := s.RuntimeError
	if s.CompileError != "" {
		errorMessage = s.CompileError
	}

	var testcase int
	var output, expected string
	for i, answer := range s.CodeAnswer {
		if answer != s.ExpectedCodeAnswer[i] {
			testcase = i
			output = answer
			expected = s.ExpectedCodeAnswer[i]
		}
	}

	isSolution := s.TaskName == SolutionTask

	return &SubmissionResult{
		Answer:            s.Correct,
		State:             s.State,
		Result:            s.StatusMsg,
		Runtime:           s.StatusRuntime,
		RuntimePercentile: fmt.Sprintf("%.2f%%", s.RuntimePercentile),
		Memory:            s.StatusMemory,
		MemoryPercentile:  fmt.Sprintf("%.2f%%", s.MemoryPercentile),
		ErrorMsg:          errorMessage,
		TestCase:          testcase,
		TestOutput:        output,
		TestExpected:      expected,
		IsSolution:        isSolution,
	}
}

type AuthResponse struct {
	Data struct {
		UserStatus UserStatus `json:"userStatus"`
	} `json:"data"`
}

type UserStatus struct {
	IsSignedIn bool   `json:"isSignedIn"`
	Username   string `json:"username"`
}

type SubmissionResult struct {
	State             string
	Answer            bool
	Result            string
	Runtime           string
	RuntimePercentile string
	Memory            string
	MemoryPercentile  string
	ErrorMsg          string
	TestCase          int
	TestOutput        string
	TestExpected      string
	IsSolution        bool
}

func (r *SubmissionResult) IsCorrect() bool {
	return r.Answer
}

func (r *SubmissionResult) HasError() bool {
	return r.ErrorMsg != ""
}
