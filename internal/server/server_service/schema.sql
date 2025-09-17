CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY NOT NULL,
    mtime INTEGER NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT NOT NULL UNIQUE COLLATE NOCASE,
    password TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE COLLATE NOCASE,
    last_presence DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    login_failed_attempts INTEGER NOT NULL DEFAULT 0 ,
    last_login_failed DATETIME DEFAULT NULL,
    totp_enabled BOOLEAN NOT NULL DEFAULT FALSE ,
    totp_secret TEXT DEFAULT NULL,
    email_otp_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    email_otp_secret TEXT DEFAULT NULL,
    photo BLOB,
    added_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS user_collaborator (
    id TEXT PRIMARY KEY NOT NULL,
    mtime INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    collaborator_id TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (collaborator_id) REFERENCES user(id),
    UNIQUE (user_id, collaborator_id),
    CHECK( typeof(collaborator_id)='text' AND length(collaborator_id)>=1),
    CHECK( collaborator_id != user_id )
);

CREATE TABLE IF NOT EXISTS studio (
    id TEXT PRIMARY KEY NOT NULL,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE COLLATE NOCASE,
    url TEXT NOT NULL DEFAULT '',
    alt_url TEXT NOT NULL DEFAULT '',
    active BOOLEAN NOT NULL DEFAULT FALSE,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    key TEXT NOT NULL UNIQUE,
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TABLE IF NOT EXISTS studio_user (
    id TEXT PRIMARY KEY NOT NULL,
    mtime INTEGER NOT NULL,
    user_id TEXT NOT NULL,
    studio_id TEXT NOT NULL,
    role_id TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(id),
    FOREIGN KEY (studio_id) REFERENCES studio(id),
    FOREIGN KEY (role_id) REFERENCES role(id),
    UNIQUE (user_id, studio_id)
);

CREATE TABLE IF NOT EXISTS role (
    id TEXT PRIMARY KEY NOT NULL,
    mtime INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE
);
