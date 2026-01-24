-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS courses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(120) UNIQUE NOT NULL,
    title       VARCHAR(255) NOT NULL,
    subtitle    VARCHAR(300),
    description TEXT,
    cover_url   VARCHAR(512),
    status      VARCHAR(30) DEFAULT 'draft',
    price       DECIMAL(10,2),
    currency    VARCHAR(3) DEFAULT 'EUR',
    level       VARCHAR(30),
    created_id UUID REFERENCES users(id),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS courses_title_idx ON courses USING GIST (title gist_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS courses;
DROP INDEX IF EXISTS courses_title_idx;
-- +goose StatementEnd
