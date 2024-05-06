DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_type WHERE typname = 'ROLE'
    ) THEN
        CREATE TYPE ROLE AS ENUM ('user', 'moderator', 'admin');
    END IF;
END $$;


CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) DEFAULT '',
    surname VARCHAR(255) DEFAULT '',
    username VARCHAR(255) NOT NULL,
    pass_hash BYTEA NOT NULL,
    phone_number VARCHAR(255) DEFAULT '',
    email VARCHAR(255) NOT NULL,
    role ROLE DEFAULT 'user' CHECK (role IN ('user', 'moderator', 'admin')),
    image_s3_path VARCHAR(255) DEFAULT '',
    is_blocked BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users_groups (
    id BIGINT PRIMARY KEY,
    user_id UUID NOT NULL,
    group_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
);

CREATE OR REPLACE FUNCTION update_modified_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_modified_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_modified_at();