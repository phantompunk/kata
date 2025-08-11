package leetcode

import (
	"testing"

	"github.com/phantompunk/kata/internal/assert"
	"github.com/phantompunk/kata/internal/models"
)

func TestLCPing(t *testing.T) {
	lc := newTestService()

	t.Run("Authenticated", func(t *testing.T) {
		loggedIn := `{"data": {"streakCounter": {"streakCount": 0}}}`
		mockResponse(loggedIn, lc.client)

		got, err := lc.Ping()
		assert.NilError(t, err)
		assert.True(t, got)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		loggedOut := `{"data": {"streakCounter": null}}`
		mockResponse(loggedOut, lc.client)

		got, err := lc.Ping()
		assert.NilError(t, err)
		assert.False(t, got)
	})
}

func TestFetch(t *testing.T) {
	lc := newTestService()

	t.Run("Problem not found", func(t *testing.T) {
		notFound := `{"data":{"question":null}}`
		mockResponse(notFound, lc.client)

		question, err := lc.Fetch("two-sum")
		assert.Equal(t, err, ErrQuestionNotFound)
		assert.Equal(t, question, nil)
	})

	t.Run("Problem found", func(t *testing.T) {
		found := `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","codeSnippets":[{"langSlug":"golang","code":"func twoSum(nums []int, target int) []int {\n    \n}"}]}}}`
		mockResponse(found, lc.client)

		question, err := lc.Fetch("two-sum")
		assert.NilError(t, err)
		assert.Equal(t, question.Title, "Two Sum")
	})
}

func TestSolutionTest(t *testing.T) {
	lc := newTestService()
	question := &models.Question{ID: "1", TitleSlug: "two-sum"}
	language := "golang"

	t.Run("Problem ", func(t *testing.T) {
		processing := `{"interpret_id":"runcode_123.456_789","test_case":"[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6"}`
		mockResponse(processing, lc.client)

		callbackUrl, err := lc.Test(question, language, "func twoSum(a int){}")
		assert.NilError(t, err)
		assert.Equal(t, callbackUrl, "https://leetcode.com/submissions/detail/runcode_123.456_789/check/")
	})
}

func TestCheckTestStatus(t *testing.T) {
	lc := newTestService()

	t.Run("Pending status", func(t *testing.T) {
		pending := `{"state": "PENDING"}`
		mockResponse(pending, lc.client)

		response, err := lc.CheckTestStatus("https://leetcode.com/submission/run123.456_789/check")
		assert.NilError(t, err)
		assert.Equal(t, response.State, "PENDING")
	})

	t.Run("Completed but failed", func(t *testing.T) {
		failed := `{"run_success":true,"correct_answer":false,"state":"SUCCESS"}`
		mockResponse(failed, lc.client)

		response, err := lc.CheckTestStatus("https://leetcode.com/submission/run123.456_789/check")
		assert.NilError(t, err)
		assert.Equal(t, response.State, "SUCCESS")
		assert.False(t, response.Correct)
	})

	t.Run("Completed and passed", func(t *testing.T) {
		failed := `{"run_success":true,"correct_answer":true,"state":"SUCCESS"}`
		mockResponse(failed, lc.client)

		response, err := lc.CheckTestStatus("https://leetcode.com/submission/run123.456_789/check")
		assert.NilError(t, err)
		assert.Equal(t, response.State, "SUCCESS")
		assert.True(t, response.Correct)
	})
}
