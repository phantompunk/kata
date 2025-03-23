package renderer

import (
	"bytes"
	"testing"

	approvals "github.com/approvals/go-approval-tests"
	"github.com/phantompunk/kata/internal/models"
)

func TestRender(t *testing.T) {
	render := New()
	problem := models.Problem{
		Content:      "Test Problem",
		TitleSlug:    "two-sum",
		FunctionName: "twoSum",
	}

	t.Run("Render Go test file", func(t *testing.T) {
		problem.LangSlug = "go"
		buf := bytes.Buffer{}
		render.Render(&buf, &problem, "test")
		approvals.VerifyString(t, buf.String())
	})

	t.Run("Render Python test file", func(t *testing.T) {
		problem.LangSlug = "python3"
		buf := bytes.Buffer{}
		render.Render(&buf, &problem, "test")
		approvals.VerifyString(t, buf.String())
	})
}
