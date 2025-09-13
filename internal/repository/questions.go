package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/phantompunk/kata/internal/models"
)

func (q *Queries) GetAllWithStatus(ctx context.Context, languages []string) ([]models.Question, error) {
	listAllWithStatuses := buildSelectClause(languages) + buildFromClause(languages)
	rows, err := q.db.QueryContext(ctx, listAllWithStatuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		q.LangStatus = make(map[string]bool)
		solvedValues := make([]int, len(languages))
		scanArgs := []any{&q.ID, &q.Title, &q.Difficulty}

		for i := range languages {
			scanArgs = append(scanArgs, &solvedValues[i])
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, err
		}

		for i, lang := range languages {
			q.LangStatus[lang] = solvedValues[i] == 1
		}
		questions = append(questions, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func (q *Question) ToModelQuestion() (*models.Question, error) {
	var modelQuestion models.Question
	modelQuestion.ID = fmt.Sprintf("%d", q.Questionid)
	modelQuestion.Title = q.Title
	modelQuestion.TitleSlug = q.Titleslug
	modelQuestion.Difficulty = q.Difficulty
	modelQuestion.FunctionName = q.Functionname
	modelQuestion.Content = q.Content

	if err := json.Unmarshal([]byte(q.Codesnippets), &modelQuestion.CodeSnippets); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(q.Testcases), &modelQuestion.TestCaseList); err != nil {
		return nil, err
	}
	return &modelQuestion, nil
}

func (q *Question) ToProblem(workspace, language string) *models.Problem {
	var problem models.Problem
	problem.QuestionID = fmt.Sprintf("%d", q.Questionid)
	problem.Content = q.Content
	problem.FunctionName = q.Functionname
	problem.TitleSlug = q.Titleslug
	problem.Slug = formatTitleSlug(q.Titleslug)
	problem.TestCases = q.Testcases

	var codeSnippets []models.CodeSnippet
	if err := json.Unmarshal([]byte(q.Codesnippets), &codeSnippets); err != nil {
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
	selectClause := "SELECT q.questionId, q.title, q.difficulty"
	for _, lang := range languages {
		selectClause += fmt.Sprintf(", COALESCE(%s.solved, 0) AS %sSolved", lang, lang)
	}
	return selectClause
}

func buildFromClause(languages []string) string {
	fromClause := " FROM questions q"
	for _, language := range languages {
		lang := strings.ToLower(language)
		fromClause += fmt.Sprintf(" LEFT JOIN submissions %s ON q.questionId = %s.questionId AND %s.langSlug = '%s'", lang, lang, lang, lang)
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
