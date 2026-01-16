ALTER TABLE users
ADD COLUMN updated_at timestamptz;

UPDATE users
SET updated_at = NOW()
WHERE updated_at IS NULL;

ALTER TABLE users
ALTER COLUMN updated_at SET NOT NULL,
ALTER COLUMN updated_at SET DEFAULT NOW();