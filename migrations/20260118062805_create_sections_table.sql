-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sections (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    created_id   UUID REFERENCES users(id),
    title           VARCHAR(255) NOT NULL,
    slug        VARCHAR(120) UNIQUE NOT NULL,
    "order"         INTEGER DEFAULT 0,
    is_free_preview BOOLEAN DEFAULT FALSE,
    estimated_time  INTEGER,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sections;
-- +goose StatementEnd
