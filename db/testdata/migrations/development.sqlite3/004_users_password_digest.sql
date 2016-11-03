-- +migrate Up
ALTER TABLE users ADD COLUMN password_digest TEXT NOT NULL DEFAULT('password');


