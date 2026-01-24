-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS lessons (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id    UUID NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    course_id     UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_id UUID REFERENCES users(id),
    title         VARCHAR(255) NOT NULL,
    slug        VARCHAR(120) UNIQUE NOT NULL,
    content       TEXT NOT NULL,
    type          VARCHAR(50),
    "order"       INTEGER DEFAULT 0,
    duration_sec  INTEGER,
    is_published  BOOLEAN DEFAULT FALSE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS lessons_section_id_idx ON lessons USING GIST (title gist_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS lessons;
DROP INDEX IF EXISTS lessons_section_id_idx;
-- +goose StatementEnd
