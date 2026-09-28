DROP TRIGGER IF EXISTS comments_updated_at ON comments;
DROP TRIGGER IF EXISTS anime_updated_at ON anime;
DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS watch_history;
DROP TABLE IF EXISTS ratings;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS anime;
DROP TABLE IF EXISTS users;