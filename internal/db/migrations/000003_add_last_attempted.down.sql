CREATE TABLE submissions_temp (
  id INTEGER PRIMARY KEY,
  questionId INTEGER,
  langSlug TEXT,
  solved INTEGER,
);

INSERT INTO submissions_temp (id, questionId, langSlug, solved)
SELECT id, questionId, langSlug, solved
FROM submissions;

DROP TABLE submissions;

ALTER TABLE submissions_temp RENAME TO submissions;
