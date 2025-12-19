-- 1. Create new table without submit_id
CREATE TABLE questions_new (
  question_id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  title_slug TEXT UNIQUE NOT NULL,
  difficulty TEXT CHECK (difficulty IN ('Easy', 'Medium', 'Hard')) NOT NULL,
  function_name TEXT NOT NULL,
  content TEXT NOT NULL,
  code_snippets TEXT NOT NULL,
  test_cases TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (DATE('now'))
);

-- 2. Copy data over (excluding submit_id)
INSERT INTO questions_new (
  question_id,
  title,
  title_slug,
  difficulty,
  function_name,
  content,
  code_snippets,
  test_cases,
  created_at
)
SELECT
  question_id,
  title,
  title_slug,
  difficulty,
  function_name,
  content,
  code_snippets,
  test_cases,
  created_at
FROM questions;

-- 3. Drop old table
DROP TABLE questions;

-- 4. Rename new table
ALTER TABLE questions_new RENAME TO questions;
