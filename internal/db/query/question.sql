-- name: GetBySlug :one
SELECT * FROM questions
WHERE titleSlug = ? LIMIT 1;

-- name: GetByID :one
SELECT * FROM questions
WHERE questionId = ? LIMIT 1;

-- name: ListAll :many
SELECT * FROM questions
ORDER BY questionId ASC;

-- name: Create :one
INSERT INTO questions (
  questionId, title, titleSlug, difficulty, functionName, content, codeSnippets, testCases
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetRandom :one
SELECT * FROM questions
ORDER BY RANDOM() LIMIT 1;

-- name: Exists :one
SELECT EXISTS (
  SELECT 1 FROM questions
  WHERE titleSlug = ?
);
