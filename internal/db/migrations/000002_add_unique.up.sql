DELETE FROM submissions
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id
        FROM submissions
        GROUP BY question_id, lang_slug
        HAVING MAX(last_attempted)
    )
);

CREATE UNIQUE INDEX idx_submissions_question_lang_unique
ON submissions(question_id, lang_slug);
