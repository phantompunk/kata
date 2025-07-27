package models

const (
	queryExists = `SELECT EXISTS(SELECT 1 FROM questions WHERE titleSlug = ?)`
	queryGetBySlug = `SELECT questionId, title, titleSlug, content, difficulty, functionName, codeSnippets FROM questions WHERE titleSlug = ?`
	queryInsert = `INSERT OR REPLACE INTO questions (questionId, title, titleSlug, difficulty, functionName, content, codeSnippets) VALUES (?, ?, ?, ?, ?, ?, ?);`
	queryGetRandom = `SELECT questionId, title, titleSlug, difficulty, content, codeSnippets FROM questions ORDER BY RANDOM() LIMIT 1;`
)
