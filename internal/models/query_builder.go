package models

import "fmt"

func buildSelectClause(languages []string) string {
	selectClause := "SELECT q.questionId, q.title, q.difficulty"
	for _, lang := range languages {
		selectClause += fmt.Sprintf(", COALESCE(%s.solved, 0) AS %sSolved", lang, lang)
	}
	return selectClause
}

func buildFromClause(languages []string) string {
	fromClause := " FROM questions q"
	for _, lang := range languages {
		fromClause += fmt.Sprintf(" LEFT JOIN status %s ON q.questionId = %s.questionId AND %s.langSlug = '%s'", lang, lang, lang, lang)
	}
	return fromClause
}
