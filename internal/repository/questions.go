package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/phantompunk/kata/internal/leetcode"
	"github.com/phantompunk/kata/internal/models"
)

func (q *Queries) GetAllWithStatus(ctx context.Context, languages []string) ([]models.QuestionStat, error) {
	listAllWithStatuses := buildSelectClause(languages) + buildFromClause(languages)
	rows, err := q.db.QueryContext(ctx, listAllWithStatuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.QuestionStat
	for rows.Next() {
		var i models.QuestionStat
		i.LangStatus = make(map[string]bool)
		solvedValues := make([]int, len(languages))
		scanArgs := []any{&i.ID, &i.Title, &i.Difficulty}

		for i := range languages {
			scanArgs = append(scanArgs, &solvedValues[i])
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for idx, lang := range languages {
			i.LangStatus[lang] = solvedValues[idx] == 1
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (q *Question) ToModelQuestion() (*models.Question, error) {
	var modelQuestion models.Question
	modelQuestion.ID = fmt.Sprintf("%d", q.QuestionID)
	modelQuestion.Title = q.Title
	modelQuestion.TitleSlug = q.TitleSlug
	modelQuestion.Difficulty = q.Difficulty
	modelQuestion.FunctionName = q.FunctionName
	modelQuestion.Content = q.Content

	if err := json.Unmarshal([]byte(q.CodeSnippets), &modelQuestion.CodeSnippets); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(q.TestCases), &modelQuestion.TestCaseList); err != nil {
		return nil, err
	}
	return &modelQuestion, nil
}

func (q *Question) ToProblem(workspace, language string) *models.Problem {
	var problem models.Problem
	problem.QuestionID = fmt.Sprintf("%d", q.QuestionID)
	problem.Content = q.Content
	problem.FunctionName = q.FunctionName
	problem.TitleSlug = q.TitleSlug
	problem.Slug = formatTitleSlug(q.TitleSlug)
	problem.TestCases = q.TestCases

	var codeSnippets []models.CodeSnippet
	if err := json.Unmarshal([]byte(q.CodeSnippets), &codeSnippets); err != nil {
		fmt.Println("Failed to unmarshal code snippets:", err)
		return nil
	}

	problem.Code = ""
	for _, snippet := range codeSnippets {
		if snippet.LangSlug == language {
			problem.Code = snippet.Code
			break
		}
	}
	problem.LangSlug = models.LangName[language]
	problem.SetPaths(workspace)
	return &problem
}

func buildSelectClause(languages []string) string {
	selectClause := "SELECT q.question_id, q.title, q.difficulty"
	for _, lang := range languages {
		selectClause += fmt.Sprintf(", COALESCE(%s.solved, 0) AS %sSolved", lang, lang)
	}
	return selectClause
}

func buildFromClause(languages []string) string {
	fromClause := " FROM questions q"
	for _, language := range languages {
		lang := strings.ToLower(language)
		fromClause += fmt.Sprintf(" LEFT JOIN submissions %s ON q.question_id = %s.question_id AND %s.lang_slug = '%s'", lang, lang, lang, lang)
	}
	return fromClause
}

var numberToString = map[string]string{"1": "one", "2": "two", "3": "three", "4": "four", "5": "five", "6": "six", "7": "seven", "8": "eight", "9": "nine", "0": "zero"}

func convertNumberToWritten(name string) string {
	letters := strings.Split(name, "")
	for i, letter := range letters {
		if hasNumber(letter) {
			written := numberToString[letter]
			letters[i] = written
		}
	}
	return strings.Join(letters, "")
}

func formatTitleSlug(name string) string {
	if hasNumber(name) {
		return convertNumberToWritten(name)
	}
	return strings.ReplaceAll(name, "-", "_")
}

func hasNumber(name string) bool {
	for _, char := range name {
		if '0' <= char && char <= '9' {
			return true
		}
	}
	return false
}

func (q *GetRandomRow) ToProblem(workspace, language string) *models.Problem {
	var problem models.Problem
	problem.QuestionID = fmt.Sprintf("%d", q.QuestionID)

	problem.TitleSlug = q.TitleSlug
	problem.Slug = formatTitleSlug(q.TitleSlug)
	problem.LastAttempted = q.LastAttempted
	problem.LangSlug = models.LangName[language]
	problem.SetPaths(workspace)
	return &problem
}

func (q *GetRandomRow) GetSolutionPath(workspace, language string) string {
	var problem models.Problem
	problem.TitleSlug = q.TitleSlug
	problem.Slug = formatTitleSlug(q.TitleSlug)
	problem.LangSlug = models.LangName[language]
	problem.SetPaths(workspace)
	return problem.SolutionPath
}

func ToRepoCreateParams(question *leetcode.Question) CreateParams {
	var params CreateParams
	qId, _ := strconv.ParseInt(question.ID, 10, 64)
	params.QuestionID = qId
	params.Title = question.Title
	params.TitleSlug = question.TitleSlug
	params.Difficulty = question.Difficulty
	params.FunctionName = question.Metadata.Name
	params.Content = question.Content

	codeSnippets, err := json.Marshal(question.CodeSnippets)
	if err != nil {
		return params
	}
	params.CodeSnippets = string(codeSnippets)

	testCases := strings.Join(question.TestCaseList, "\n")
	params.TestCases = string(testCases)

	return params
}
