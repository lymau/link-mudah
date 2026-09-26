CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE page_settings (
    user_id UUID NOT NULL UNIQUE REFERENCES users (id),
    bg_color TEXT,
    font_family TEXT,
    avatar_url TEXT,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE links (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id),
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    position INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);