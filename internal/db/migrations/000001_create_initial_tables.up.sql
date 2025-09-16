CREATE TABLE questions (
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

CREATE TABLE submissions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  question_id INTEGER NOT NULL,
  lang_slug TEXT NOT NULL,
  solved INTEGER CHECK (solved IN (0, 1)) NOT NULL,
  last_attempted TEXT NOT NULL DEFAULT (DATE('now')),
  FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);
