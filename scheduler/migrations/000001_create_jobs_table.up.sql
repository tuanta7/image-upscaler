CREATE TABLE jobs (
    id         TEXT PRIMARY KEY,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
