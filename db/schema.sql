\c social_net;
SET ROLE TO admin_user;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(50) NOT NULL,
    second_name VARCHAR(50) NOT NULL,
    birthdate DATE NOT NULL,
    city VARCHAR(50) NOT NULL,
    biography VARCHAR(255)
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
    user_id UUID NOT NULL,
    friend_id UUID NOT NULL
);

/*CREATE TABLE IF NOT EXISTS dialogs (
    author_id UUID NOT NULL, 
    recepient_id UUID NOT NULL,
    time TIMESTAMP,
    message VARCHAR(400)
);*/

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY, /* post id */
    author_id UUID NOT NULL /* author_id */
);

CREATE INDEX IF NOT EXISTS users_idx ON users(first_name, second_name);