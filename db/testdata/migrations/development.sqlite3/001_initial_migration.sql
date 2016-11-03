-- +migrate Up
-- Initial migration, articles
CREATE TABLE IF NOT EXISTS ARTICLES (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    created TIMESTAMP DEFAULT( DATETIME('now') ), 
    updated TIMESTAMP DEFAULT( DATETIME('now') )
);

-- +migrate Down
-- remove articles
DROP TABLE ARTICLES;