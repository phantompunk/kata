CREATE TABLE questions (
  questionId INTEGER PRIMARY KEY,
  title TEXT NOT NULL,
  titleSlug TEXT UNIQUE NOT NULL,
  difficulty TEXT CHECK (difficulty IN ('Easy', 'Medium', 'Hard')) NOT NULL,
  functionName TEXT NOT NULL,
  content TEXT NOT NULL,
  codeSnippets TEXT NOT NULL
);

CREATE TABLE submissions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  questionId INTEGER NOT NULL,
  langSlug TEXT NOT NULL,
  solved INTEGER CHECK (solved IN (0, 1)) NOT NULL,
  FOREIGN KEY (questionId) REFERENCES questions(questionId) ON DELETE CASCADE
);
