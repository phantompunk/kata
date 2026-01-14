-- SQLite doesn't support DROP COLUMN, recreate table without new columns
CREATE TABLE submissions_backup (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  question_id INTEGER NOT NULL,
  lang_slug TEXT NOT NULL,
  solved INTEGER CHECK (solved IN (0, 1)) NOT NULL,
  last_attempted TEXT NOT NULL DEFAULT (DATE('now')),
  FOREIGN KEY (question_id) REFERENCES questions(question_id) ON DELETE CASCADE
);

INSERT INTO submissions_backup (id, question_id, lang_slug, solved, last_attempted)
SELECT id, question_id, lang_slug, solved, last_attempted FROM submissions;

DROP TABLE submissions;

ALTER TABLE submissions_backup RENAME TO submissions;

CREATE UNIQUE INDEX idx_submissions_question_lang ON submissions(question_id, lang_slug);
