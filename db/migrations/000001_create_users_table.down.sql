-- 000001_create_users_table.down.sql
-- Drops the users table

DROP INDEX IF EXISTS idx_users_email;
DROP TABLE IF EXISTS users;
