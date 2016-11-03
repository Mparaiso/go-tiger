-- +migrate Up
ALTER TABLE articles ADD COLUMN author_id INTEGER REFERENCES users(id) ON DELETE SET NULL ;