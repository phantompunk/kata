package models

import (
	"testing"

	"github.com/phantompunk/kata/internal/assert"
)

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
