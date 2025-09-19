-- name: GetBySlug :one
SELECT * FROM questions
WHERE title_slug = ? LIMIT 1;

-- name: GetByID :one
SELECT * FROM questions
WHERE question_id = ? LIMIT 1;

-- name: ListAll :many
SELECT * FROM questions
ORDER BY question_id ASC;

-- name: Create :one
INSERT INTO questions (
  question_id, title, title_slug, difficulty, function_name, content, code_snippets, test_cases, created_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, DATE('now')
) RETURNING *;

-- name: GetRandom :one
SELECT q.question_id, q.title, q.title_slug, q.difficulty, COALESCE(s.solved, 0) AS solved,
    COALESCE(s.last_attempted, q.created_at) AS last_attempted
FROM questions q
LEFT JOIN submissions s ON s.question_id = q.question_id
WHERE q.question_id = (
    SELECT question_id FROM questions
    ORDER BY RANDOM()
    LIMIT 1
) LIMIT 1;

-- name: Exists :one
SELECT EXISTS (
  SELECT 1 FROM questions
  WHERE title_slug = ?
);

-- name: Submit :one
INSERT INTO submissions (
  id, question_id, lang_slug, solved, last_attempted
) VALUES (
  ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetStats :one
SELECT
    COUNT(DISTINCT q.question_id) AS attempted,
    COUNT(DISTINCT CASE WHEN s.solved = 1 THEN s.question_id END) AS completed
FROM questions q
LEFT JOIN submissions s on q.question_id = s.question_id;
