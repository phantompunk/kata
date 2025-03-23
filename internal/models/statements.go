package models

var gQLQueryQuestion string = `query questionEditorData($titleSlug: String!) {
  question(titleSlug: $titleSlug) {
    questionId
    content
    titleSlug
		title
		difficulty
    codeSnippets {
      langSlug
      code
    }
  }
}`

var queryGetBySlug string = `SELECT
	questionId, title, titleSlug, content, difficulty, functionName, codeSnippets 
	FROM questions
	WHERE titleSlug = ?;`

var queryExists = `SELECT EXISTS(SELECT 1 FROM questions WHERE titleSlug=?);`
