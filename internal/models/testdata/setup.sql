CREATE TABLE questions (
    questionId INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    titleSlug TEXT UNIQUE NOT NULL,
    content TEXT NOT NULL,
    difficulty TEXT CHECK (difficulty IN ('Easy', 'Medium', 'Hard')) NOT NULL,
    functionName TEXT NOT NULL,
    codeSnippets TEXT NOT NULL
);

CREATE TABLE status (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    questionId INTEGER NOT NULL,
    langSlug TEXT NOT NULL,
    solved INTEGER CHECK (solved IN (0, 1)) NOT NULL,
    FOREIGN KEY (questionId) REFERENCES questions(questionId) ON DELETE CASCADE
);

INSERT INTO questions (questionId, title, titleSlug, content, difficulty, functionName, codeSnippets) VALUES (
    36,
    'Two Sum',
    'two-sum',
    'Sample problem description',
    'Easy',
    'twoSum',
    '[{"langSlug": "cpp","code": "Function Sample()"}]'
);

INSERT INTO status (questionId, langSlug, solved) VALUES
(36, 'go', 0),
(36, 'python', 0),
(36, 'java', 1);
