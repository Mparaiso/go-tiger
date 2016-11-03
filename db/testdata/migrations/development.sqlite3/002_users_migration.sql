-- +migrate Up
CREATE TABLE users(
    id INTEGER PRIMARY KEY AUTOINCREMENT ,
    name VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created TIMESTAMP DEFAULT( DATETIME('now') ), 
    updated TIMESTAMP DEFAULT( DATETIME('now') ),
    userinfo_id INTEGER
);

-- +migrate Down
DROP TABLE users;
