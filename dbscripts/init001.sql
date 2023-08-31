CREATE TABLE users (
    id bigint primary key generated always as identity,
    created_at timestamptz default NOW() NOT NULL,
    username varchar(255) unique NOT NULL,
    password varchar(255) NOT NULL,
)
