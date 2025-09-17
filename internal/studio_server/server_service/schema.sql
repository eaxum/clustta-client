CREATE TABLE IF NOT EXISTS config (
    name TEXT PRIMARY KEY NOT NULL COLLATE NOCASE,
    value CLOB,
    mtime INTEGER NOT NULL,
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TABLE IF NOT EXISTS "role" (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    name TEXT UNIQUE NOT NULL COLLATE NOCASE,

    view_project BOOLEAN DEFAULT FALSE NOT NULL,
    create_project BOOLEAN DEFAULT FALSE NOT NULL,
    update_project BOOLEAN DEFAULT FALSE NOT NULL,
    delete_project BOOLEAN DEFAULT FALSE NOT NULL,
    
    CHECK( typeof(name)='text' AND length(name)>=1)
);

CREATE TABLE IF NOT EXISTS user (
    id TEXT PRIMARY KEY,
    mtime INTEGER NOT NULL,
    added_at DATETIME NOT NULL,
    first_name TEXT COLLATE NOCASE,
    last_name TEXT COLLATE NOCASE,
    username TEXT UNIQUE COLLATE NOCASE,
    email TEXT NOT NULL UNIQUE COLLATE NOCASE,
    photo BLOB,
    role_id TEXT NOT NULL,
    FOREIGN KEY (role_id) REFERENCES role(id),
	CHECK( typeof(first_name)='text' AND length(first_name)>=1),
	CHECK( typeof(last_name)='text' AND length(last_name)>=1),
	CHECK( typeof(username)='text' AND length(username)>=1),
	CHECK( typeof(email)='text' AND length(email)>=1)
);

