\c social_net;
SET ROLE TO admin_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(50) NOT NULL,
    second_name VARCHAR(50) NOT NULL,
    birthdate DATE NOT NULL,
    city VARCHAR(50) NOT NULL,
    biography VARCHAR(255),
    is_celebrity BOOLEAN DEFAULT 0 NOT NULL
);

CREATE TABLE IF NOT EXISTS user_credentials (
    id UUID PRIMARY KEY,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_tokens (
    id UUID PRIMARY KEY,
    token TEXT NOT NULL,
    valid_until TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS friends (
    id UUID REFERENCES users(id)  NOT NULL,
    friend_id UUID NOT NULL,
    PRIMARY KEY(id, friend_id) 
);

CREATE TABLE IF NOT EXISTS posts (
    id UUID DEFAULT uuid_generate_v4(),
    author_user_id UUID NOT NULL,
    text VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    PRIMARY KEY(id, author_user_id)
);

CREATE INDEX IF NOT EXISTS users_idx ON users(first_name, second_name);