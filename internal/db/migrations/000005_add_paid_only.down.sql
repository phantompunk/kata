-- SQLite doesn't support DROP COLUMN, must recreate table
CREATE TABLE questions_new (
  question_id INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  title_slug TEXT UNIQUE NOT NULL,
  difficulty TEXT CHECK (difficulty IN ('Easy', 'Medium', 'Hard')) NOT NULL,
  function_name TEXT NOT NULL,
  content TEXT NOT NULL,
  code_snippets TEXT NOT NULL,
  test_cases TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (DATE('now')),
  submit_id INTEGER
);

INSERT INTO questions_new SELECT
  question_id, title, title_slug, difficulty, function_name,
  content, code_snippets, test_cases, created_at, submit_id
FROM questions;

DROP TABLE questions;
ALTER TABLE questions_new RENAME TO questions;
