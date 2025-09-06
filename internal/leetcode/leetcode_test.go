package leetcode

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/phantompunk/kata/internal/assert"
	"github.com/phantompunk/kata/internal/models"
)

func TestLCPing(t *testing.T) {
	lc := newTestService()

	t.Run("Authenticated", func(t *testing.T) {
		loggedIn := `{"data": {"userStatus": {"isSignedIn": true,"username": ""}}}`
		mockResponse(http.StatusOK, loggedIn, lc.client)

		got, err := lc.Ping()
		assert.NilError(t, err)
		assert.True(t, got)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		loggedOut := `{"data": {"userStatus": {"isSignedIn": false,"username": ""}}}`
		mockResponse(http.StatusOK, loggedOut, lc.client)

		got, err := lc.Ping()
		assert.NilError(t, err)
		assert.False(t, got)
	})

	t.Run("Unknown", func(t *testing.T) {
		loggedOut := `{"adata": {"userStatus": {"isSignedIn": false,"username": ""}}}`
		mockResponse(http.StatusUnauthorized, loggedOut, lc.client)

		got, err := lc.Ping()
		assert.Equal(t, err, ErrNotAuthenticated)
		assert.False(t, got)
	})
}

func TestFetch(t *testing.T) {
	lc := newTestService()

	t.Run("Problem not found", func(t *testing.T) {
		notFound := `{"data":{"question":null}}`
		mockResponse(http.StatusOK, notFound, lc.client)

		question, err := lc.Fetch("two-sum")
		assert.Equal(t, err, ErrQuestionNotFound)
		assert.Equal(t, question, nil)
	})

	t.Run("Problem found", func(t *testing.T) {
		found := `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","codeSnippets":[{"langSlug":"golang","code":"func twoSum(nums []int, target int) []int {\n    \n}"}]}}}`
		mockResponse(http.StatusOK, found, lc.client)

		question, err := lc.Fetch("two-sum")
		assert.NilError(t, err)
		assert.Equal(t, question.Title, "Two Sum")
		assert.Equal(t, question.ID, "1")
	})

	t.Run("Problem found by id", func(t *testing.T) {
		found := `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","codeSnippets":[{"langSlug":"golang","code":"func twoSum(nums []int, target int) []int {\n    \n}"}]}}}`
		mockResponse(http.StatusOK, found, lc.client)

		question, err := lc.Fetch("1")
		assert.NilError(t, err)
		assert.Equal(t, question.Title, "Two Sum")
		assert.Equal(t, question.ID, "1")
	})
}

func TestSolutionTest(t *testing.T) {
	lc := newTestService()
	question := &models.Question{ID: "1", TitleSlug: "two-sum"}
	language := "golang"

	t.Run("Problem ", func(t *testing.T) {
		processing := `{"interpret_id":"runcode_123.456_789","test_case":"[2,7,11,15]\n9\n[3,2,4]\n6\n[3,3]\n6"}`
		mockResponse(http.StatusOK, processing, lc.client)

		callbackUrl, err := lc.Test(question.ToProblem("", language), language, "func twoSum(a int){}")
		assert.NilError(t, err)
		assert.Equal(t, callbackUrl, "https://leetcode.com/submissions/detail/runcode_123.456_789/check/")
	})
}

func TestCheckTestStatus(t *testing.T) {
	lc := newTestService()

	t.Run("Pending status", func(t *testing.T) {
		pending := `{"state": "PENDING"}`
		mockResponse(http.StatusOK, pending, lc.client)

		response, err := lc.CheckTestStatus("https://leetcode.com/submission/run123.456_789/check")
		assert.NilError(t, err)
		assert.Equal(t, response.State, "PENDING")
	})

	t.Run("Completed but failed", func(t *testing.T) {
		failed := `{"run_success":true,"correct_answer":false,"state":"SUCCESS"}`
		mockResponse(http.StatusOK, failed, lc.client)

		response, err := lc.CheckTestStatus("https://leetcode.com/submission/run123.456_789/check")
		assert.NilError(t, err)
		assert.Equal(t, response.State, "SUCCESS")
		assert.False(t, response.Correct)
	})

	t.Run("Completed and passed", func(t *testing.T) {
		failed := `{"run_success":true,"correct_answer":true,"state":"SUCCESS"}`
		mockResponse(http.StatusOK, failed, lc.client)

		response, err := lc.CheckTestStatus("https://leetcode.com/submission/run123.456_789/check")
		assert.NilError(t, err)
		assert.Equal(t, response.State, "SUCCESS")
		assert.True(t, response.Correct)
	})
}
