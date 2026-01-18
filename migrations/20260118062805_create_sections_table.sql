-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sections (
    id              BIGSERIAL PRIMARY KEY,
    course_id       BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title           VARCHAR(255) NOT NULL,
    "order"           INTEGER DEFAULT 0,
    is_free_preview BOOLEAN DEFAULT FALSE,
    estimated_time  INTEGER,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS sections;
-- +goose StatementEnd
