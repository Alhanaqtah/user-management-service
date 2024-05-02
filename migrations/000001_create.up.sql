CREATE TYPE ROLE AS ENUM('user', 'moderator', 'admin');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) DEFAULT '',
    surname VARCHAR(255) DEFAULT '',
    username VARCHAR(255) NOT NULL,
    pass_hash BYTEA NOT NULL,
    phone_number VARCHAR(255) DEFAULT '',
    email VARCHAR(255) NOT NULL,
    role ROLE DEFAULT 'user' CHECK (role IN ('user', 'moderator', 'admin')),
    image_s3_path VARCHAR(255),
    is_blocked BOOLEAN DEFAULT false,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP(0) WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(0) WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users_groups (
    id BIGINT PRIMARY KEY,
    user_id UUID NOT NULL,
    group_id UUID NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
);

ALTER TABLE users_groups
    ADD CONSTRAINT users_groups_group_id_foreign FOREIGN KEY (group_id) REFERENCES groups(id);

ALTER TABLE users_groups
    ADD CONSTRAINT users_groups_user_id_foreign FOREIGN KEY (user_id) REFERENCES users(id);
