DROP IF EXISTS user_groups;
DROP IF EXISTS group;
DROP IF EXISTS users;

DROP IF EXISTS ROLE;

DROP TRIGGER IF EXISTS update_users_modified_at ON users;
DROP FUNCTION IF EXISTS update_modified_at();