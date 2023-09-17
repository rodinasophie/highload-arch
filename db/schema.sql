\c social_net;
SET ROLE TO admin_user;

CREATE EXTENSION pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    first_name VARCHAR(50) NOT NULL,
    second_name VARCHAR(50) NOT NULL,
    birthdate DATE NOT NULL,
    city VARCHAR(50) NOT NULL,
    biography VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS user_credentials (
    id UUID PRIMARY KEY ,
    password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS user_tokens(
    id UUID PRIMARY KEY,
    token TEXT NOT NULL,
    valid_until TIMESTAMP NOT NULL
);
