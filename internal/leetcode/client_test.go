package leetcode

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/phantompunk/kata/internal/domain"
	"github.com/phantompunk/kata/pkg/assert"
)

func TestFetchQuestion(t *testing.T) {
	resp := &Responder{}
	client := newTestClient(resp)
	slug := "two-sum"

	t.Run("Problem not found", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"question":null}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.Equal(t, err, ErrQuestionNotFound)
		assert.Equal(t, question, nil)
	})

	t.Run("Problem found", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"question":{"questionFrontendId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","metaData": "{\r\n  \"name\": \"twoSum\",\r\n  \"params\": [\r\n    {\r\n      \"name\": \"nums\",\r\n      \"type\": \"integer[]\"\r\n    },\r\n    {\r\n      \"name\": \"target\",\r\n      \"type\": \"integer\"\r\n    }\r\n  ],\r\n  \"return\": {\r\n    \"type\": \"list<list<integer>>\",\r\n    \"colsize\": 4,\r\n    \"dealloc\": true\r\n  }\r\n}"}}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.NilError(t, err)
		assert.Equal(t, question.ID, "1")
		assert.Equal(t, question.Title, "Two Sum")
	})

	t.Run("Problem metadata", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"question":{"questionFrontendId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","metaData": "{\r\n  \"name\": \"twoSum\",\r\n  \"params\": [\r\n    {\r\n      \"name\": \"nums\",\r\n      \"type\": \"integer[]\"\r\n    },\r\n    {\r\n      \"name\": \"target\",\r\n      \"type\": \"integer\"\r\n    }\r\n  ],\r\n  \"return\": {\r\n    \"type\": \"list<list<integer>>\",\r\n    \"colsize\": 4,\r\n    \"dealloc\": true\r\n  }\r\n}"}}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.NilError(t, err)
		assert.Equal(t, question.Metadata.Name, "twoSum")
	})
}

func TestSubmitQuestion(t *testing.T) {
	resp := &Responder{}
	client := newTestClient(resp)
	problem := &domain.Problem{ID: "1", Slug: "two-sum", Testcases: []string{"[2,7]\n9"}, Language: domain.NewProgrammingLanguage("go")}
	submissionId := "12345"
	snippet := "func twoSum(){}"

	t.Run("valid solution submission", func(t *testing.T) {
		resp.SetResponse(200, `{"submission_id": 12345,"test_case": null}`)
		id, err := client.SubmitSolution(context.Background(), problem, snippet)

		assert.NilError(t, err)
		assert.Equal(t, id, submissionId)
	})

	t.Run("Valid test submission", func(t *testing.T) {
		resp.SetResponse(200, `{"interpret_id": "12345","test_case": "[2,7]\n9"}`)
		id, err := client.SubmitTest(context.Background(), problem, snippet)

		assert.NilError(t, err)
		assert.Equal(t, id, submissionId)
	})
}

func TestCheckResult(t *testing.T) {
	resp := &Responder{}
	client := newTestClient(resp)
	submissionId := "12345"

	t.Run("Result is pending", func(t *testing.T) {
		resp.SetResponse(200, `{"state": "PENDING"}`)
		result, err := client.CheckSubmissionResult(context.Background(), submissionId)

		assert.NilError(t, err)
		assert.Equal(t, result.State, "PENDING")
	})

	t.Run("Resulted in runtime error", func(t *testing.T) {
		resp.SetResponse(200, `{"state": "SUCCESS", "status_msg": "Runtime Error","run_success": false,"runtime_error": "SyntaxError: Invalid or unexpected token"}`)
		result, err := client.CheckSubmissionResult(context.Background(), submissionId)

		assert.NilError(t, err)
		assert.Equal(t, result.State, "SUCCESS")
		assert.Equal(t, result.Result, "Runtime Error")
		assert.Equal(t, result.Answer, false)
	})

	t.Run("Submit solution", func(t *testing.T) {
		resp.SetResponse(200, `{"status_runtime":"0 ms","status_memory":"3.9 MB","status_msg":"Accepted"}`)
		result, err := client.CheckSubmissionResult(context.Background(), submissionId)

		assert.NilError(t, err)
		assert.Equal(t, result.Runtime, "0 ms")
	})
}

func TestUserStatus(t *testing.T) {
	resp := &Responder{}
	client := newTestClient(resp)

	t.Run("User logged out", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"userStatus":{"isSignedIn":false}}}`)
		authenticated, err := client.IsAuthenticated(context.Background())

		assert.NilError(t, err)
		assert.Equal(t, authenticated, false)
	})

	t.Run("User signed in", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"userStatus":{"isSignedIn":true,"username":"tester"}}}`)
		username, err := client.GetUsername(context.Background())

		assert.NilError(t, err)
		assert.Equal(t, username, "tester")
	})
}

type Responder struct {
	Status int
	Body   string
}

func (r *Responder) SetResponse(status int, body string) {
	r.Status = status
	r.Body = body
}

func (r *Responder) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: r.Status,
		Body:       io.NopCloser(bytes.NewReader([]byte(r.Body))),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func newTestClient(responder *Responder) *LeetCodeClient {
	client := &http.Client{}
	client.Transport = responder
	return WithClient(client)
}
