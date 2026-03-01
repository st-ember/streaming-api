CREATE TABLE IF NOT EXISTS videos (
    id TEXT PRIMARY KEY,
    title TEXT,
    description TEXT,
    duration BIGINT,
    filename TEXT,
    resource_id TEXT,
    status TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    video_id TEXT,
    type TEXT,
    status TEXT,
    result TEXT,
    error_msg TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);
