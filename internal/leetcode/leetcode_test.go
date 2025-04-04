package leetcode

import (
	"net/http"
	"testing"

	"github.com/phantompunk/kata/internal/assert"
)

func TestLCPing(t *testing.T) {
	mockClient := &http.Client{}
	session, csrftoken := "abc123", "csrf123"
	lc := Service{mockClient, "", session, csrftoken}

	t.Run("Authenticated", func(t *testing.T) {
		loggedIn := `{"data": {"streakCounter": {"streakCount": 0}}}`
		mockResponse(loggedIn, mockClient)

		got, err := lc.Ping()
		assert.NilError(t, err)
		assert.True(t, got)
	})

	t.Run("Unauthenticated", func(t *testing.T) {
		loggedOut := `{"data": {"streakCounter": null}}`
		mockResponse(loggedOut, mockClient)

		got, err := lc.Ping()
		assert.NilError(t, err)
		assert.False(t, got)
	})
}

func TestFetch(t *testing.T) {
	mockClient := &http.Client{}
	session, csrftoken := "abc123", "csrf123"
	lc := Service{mockClient, "", session, csrftoken}

	t.Run("Problem not found", func(t *testing.T) {
		notFound := `{"data":{"question":null}}`
		mockResponse(notFound, mockClient)

		question, err := lc.Fetch("two-sum")
		assert.Equal(t, err, ErrNotFound)
		assert.Equal(t, question, nil)
	})

	t.Run("Problem found", func(t *testing.T) {
		found := `{"data":{"question":{"questionId":"1","content":"<p>Given an array of integers</p>","titleSlug":"two-sum","title":"Two Sum","difficulty":"Easy","codeSnippets":[{"langSlug":"golang","code":"func twoSum(nums []int, target int) []int {\n    \n}"}]}}}`
		mockResponse(found, mockClient)

		question, err := lc.Fetch("two-sum")
		assert.NilError(t, err)
		assert.Equal(t, question.Title, "Two Sum")
	})
}
