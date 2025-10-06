package leetcode

import (
	"context"
	"testing"

	"github.com/phantompunk/kata/internal/assert"
	"github.com/phantompunk/kata/internal/domain"
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
		resp.SetResponse(200, `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","metaData": "{\r\n  \"name\": \"twoSum\",\r\n  \"params\": [\r\n    {\r\n      \"name\": \"nums\",\r\n      \"type\": \"integer[]\"\r\n    },\r\n    {\r\n      \"name\": \"target\",\r\n      \"type\": \"integer\"\r\n    }\r\n  ],\r\n  \"return\": {\r\n    \"type\": \"list<list<integer>>\",\r\n    \"colsize\": 4,\r\n    \"dealloc\": true\r\n  }\r\n}"}}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.NilError(t, err)
		assert.Equal(t, question.ID, "1")
		assert.Equal(t, question.Title, "Two Sum")
	})

	t.Run("Problem metadata", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","metaData": "{\r\n  \"name\": \"twoSum\",\r\n  \"params\": [\r\n    {\r\n      \"name\": \"nums\",\r\n      \"type\": \"integer[]\"\r\n    },\r\n    {\r\n      \"name\": \"target\",\r\n      \"type\": \"integer\"\r\n    }\r\n  ],\r\n  \"return\": {\r\n    \"type\": \"list<list<integer>>\",\r\n    \"colsize\": 4,\r\n    \"dealloc\": true\r\n  }\r\n}"}}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.NilError(t, err)
		assert.Equal(t, question.Metadata.Name, "twoSum")
	})
}

func TestSubmitQuestion(t *testing.T) {
	resp := &Responder{}
	client := newTestClient(resp)
	problem := &domain.Problem{ID: "1", Slug: "two-sum", Testcases: "[2,7]\n9", Language: domain.NewProgrammingLanguage("go")}
	submissionId := "12345"
	snippet := "func twoSum(){}"

	t.Run("Submit solution", func(t *testing.T) {
		resp.SetResponse(200, `{"interpret_id": "12345","test_case": null}`)
		id, err := client.SubmitSolution(context.Background(), problem, snippet)

		assert.NilError(t, err)
		assert.Equal(t, id, submissionId)
	})

	t.Run("Submit test", func(t *testing.T) {
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

	t.Run("Submit solution", func(t *testing.T) {
		resp.SetResponse(200, `{"status_runtime":"0 ms","status_memory":"3.9 MB","status_msg":"Accepted"}`)
		result, err := client.CheckSubmissionResult(context.Background(), submissionId)

		assert.NilError(t, err)
		assert.Equal(t, result.Runtime, "0 ms")
	})
}
