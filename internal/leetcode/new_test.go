package leetcode

import (
	"context"
	"testing"

	"github.com/phantompunk/kata/internal/assert"
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
		resp.SetResponse(200, `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","codeSnippets":[{"langSlug":"golang","code":"func twoSum(nums []int, target int) []int {\n    \n}"}],"exampleTestcaseList":["Input: nums = [2,7,11,15], target = 9\nOutput: [0,1]"]}}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.NilError(t, err)
		assert.Equal(t, question.ID, "1")
		assert.Equal(t, question.Title, "Two Sum")
	})

	t.Run("Problem metadata", func(t *testing.T) {
		resp.SetResponse(200, `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","metaData": "{\r\n  \"name\": \"twoSum\",\r\n  \"params\": [\r\n    {\r\n      \"name\": \"nums\",\r\n      \"type\": \"integer[]\"\r\n    },\r\n    {\r\n      \"name\": \"target\",\r\n      \"type\": \"integer\"\r\n    }\r\n  ],\r\n  \"return\": {\r\n    \"type\": \"list<list<integer>>\",\r\n    \"colsize\": 4,\r\n    \"dealloc\": true\r\n  }\r\n}"}}}`)
		question, err := client.FetchQuestion(context.Background(), slug)

		assert.NilError(t, err)
		assert.Equal(t, question.Metadata.Name, "fourSum")
	})
}
