-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS lessons (
    id            BIGSERIAL PRIMARY KEY,
    section_id    BIGINT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    title         VARCHAR(255) NOT NULL,
    type          VARCHAR(50),
    "order"         INTEGER DEFAULT 0,
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
