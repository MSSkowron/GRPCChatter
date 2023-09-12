CREATE TABLE roles (
    id bigint primary key generated always as identity,
    name varchar(255) unique NOT NULL,
);

INSERT INTO roles (name) VALUES ('USER');
INSERT INTO roles (name) VALUES ('ADMIN');

ALTER TABLE users ADD COLUMN role_id bigint REFERENCES roles(id) default (1);

UPDATE users SET role_id = 1;