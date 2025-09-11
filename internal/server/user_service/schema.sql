
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    last_presence DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    login_failed_attempts INTEGER NOT NULL DEFAULT 0 ,
    last_login_failed DATETIME DEFAULT NULL,
    totp_enabled BOOLEAN NOT NULL DEFAULT FALSE ,
    totp_secret TEXT DEFAULT NULL,
    email_otp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    email_otp_secret TEXT DEFAULT NULL,
    photo BLOB,
    added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT FALSE

);
-- CREATE TABLE IF NOT EXISTS login_log (
--     id TEXT PRIMARY KEY NOT NULL,
--     last_login DATETIME NOT NULL 
-- );