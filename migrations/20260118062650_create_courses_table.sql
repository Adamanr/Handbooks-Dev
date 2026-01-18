-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS courses (
    id          BIGSERIAL PRIMARY KEY,
    slug        VARCHAR(120) UNIQUE NOT NULL,
    title       VARCHAR(255) NOT NULL,
    subtitle    VARCHAR(300),
    description TEXT,
    cover_url   VARCHAR(512),
    status      VARCHAR(30) DEFAULT 'draft',
    price       DECIMAL(10,2),
    currency    VARCHAR(3) DEFAULT 'EUR',
    level       VARCHAR(30),
    created_by_id BIGINT REFERENCES users(id),
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_courses_slug ON courses(slug);
CREATE INDEX idx_courses_status ON courses(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS courses;
-- +goose StatementEnd
