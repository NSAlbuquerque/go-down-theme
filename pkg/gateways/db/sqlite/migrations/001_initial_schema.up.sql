CREATE TABLE themes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT DEFAULT '',
    url TEXT NOT NULL,
    hash TEXT DEFAULT '',             -- tmTheme file MD5 hash.
    light BOOLEAN DEFAULT 0,
    version TEXT DEFAULT '',

    project_repo_id TEXT DEFAULT '', -- ID generated from the MD5 hash of the URL.
    project_repo TEXT DEFAULT '',
    readme TEXT DEFAULT '',
    license TEXT DEFAULT '',
    provider TEXT NOT NULL,
    last_update DATETIME, -- last project update.
    
    created_at DATETIME DEFAULT (strftime('%s', 'now')),
    updated_at DATETIME
);