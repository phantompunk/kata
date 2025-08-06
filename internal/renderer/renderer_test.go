package renderer

import (
	"bytes"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/phantompunk/kata/internal/assert"
	"github.com/phantompunk/kata/internal/models"
)

func TestNewRenderer(t *testing.T) {
	r, err := New()
	assert.NilError(t, err)
	assert.NotNil(t, r)
}

func TestRenderTemplates(t *testing.T) {
	r, _ := New()
	problem := models.Problem{
		Content:      "Test Problem",
		TitleSlug:    "two-sum",
		FunctionName: "twoSum",
		LangSlug:     "go",
		Code:         "func twoSum(nums []int, target int) []int {\n    \n}",
	}

	t.Run("Render solution file", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := r.Render(&buf, &problem, "solution")
		assert.NilError(t, err)
		approvals.VerifyString(t, buf.String())
	})

	t.Run("Render test file", func(t *testing.T) {
		buf := bytes.Buffer{}
		err := r.Render(&buf, &problem, "test")
		assert.NilError(t, err)
		approvals.VerifyString(t, buf.String())
	})

	t.Run("Render Python solution file", func(t *testing.T) {
		problem.LangSlug = "python3"
		problem.Code = "class Solution:\n    def twoSum(self, nums: List[int], target: int) -> List[int]:\n"

		buf := bytes.Buffer{}
		r.Render(&buf, &problem, "solution")
		approvals.VerifyString(t, buf.String())
	})

	t.Run("Render Python test file", func(t *testing.T) {
		problem.LangSlug = "python3"
		buf := bytes.Buffer{}
		r.Render(&buf, &problem, "test")
		approvals.VerifyString(t, buf.String())
	})
}
