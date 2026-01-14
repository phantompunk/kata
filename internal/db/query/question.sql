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
  question_id, submit_id, title, title_slug, difficulty, function_name, content, code_snippets, test_cases, created_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?, ?, ?, ? 
) ON CONFLICT(question_id) DO UPDATE SET
    title       = excluded.title,
    title_slug  = excluded.title_slug,
    difficulty  = excluded.difficulty,
    created_at  = excluded.created_at
RETURNING *;

-- name: GetRandom :one
SELECT q.question_id, q.title, q.title_slug, q.difficulty,
  CASE
      WHEN s.solved = 1 THEN 'Completed'
      ELSE 'Attempted'
  END AS status,
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
  question_id, lang_slug, solved, last_attempted
) VALUES (
  ?, ?, ?, ? 
) ON CONFLICT(question_id, lang_slug) DO UPDATE SET
    solved  = excluded.solved,
    last_attempted  = excluded.last_attempted
RETURNING *;

-- name: GetStats :one
SELECT
    COUNT(DISTINCT q.question_id) AS attempted,
    COUNT(DISTINCT CASE WHEN s.solved = 1 THEN s.question_id END) AS completed
FROM questions q
LEFT JOIN submissions s on q.question_id = s.question_id;

-- name: GetRandomWeighted :one
SELECT q.question_id, q.title, q.title_slug, q.difficulty,
  CASE WHEN s.solved = 1 THEN 'Completed' ELSE 'Attempted' END AS status,
  COALESCE(s.last_attempted, q.created_at) AS last_attempted,
  (
    CAST(julianday('now') - julianday(COALESCE(s.last_attempted, q.created_at)) AS REAL) * 0.4 +
    COALESCE(s.failed_attempts, 0) * 0.3 +
    CASE q.difficulty WHEN 'Easy' THEN 0.3 WHEN 'Medium' THEN 0.6 ELSE 1.0 END * 0.2 +
    CASE WHEN COALESCE(s.times_solved, 0) = 1 THEN 0.1 ELSE 0.0 END
  ) AS weight_score
FROM questions q
LEFT JOIN submissions s ON s.question_id = q.question_id
ORDER BY weight_score DESC, RANDOM()
LIMIT 1;

-- name: IncrementFailedAttempts :exec
UPDATE submissions
SET failed_attempts = failed_attempts + 1, last_attempted = ?
WHERE question_id = ? AND lang_slug = ?;

-- name: IncrementTimesSolved :exec
UPDATE submissions
SET times_solved = times_solved + 1, solved = 1, last_attempted = ?
WHERE question_id = ? AND lang_slug = ?;
