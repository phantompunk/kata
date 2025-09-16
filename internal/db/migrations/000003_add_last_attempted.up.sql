ALTER TABLE submissions
ADD COLUMN lastAttempted TEXT;

UPDATE submissions
SET lastAttempted = DATE('now');
