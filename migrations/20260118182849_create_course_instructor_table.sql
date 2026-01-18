-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS course_instructors(
    id              BIGSERIAL PRIMARY KEY,
    course_id       BIGINT NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_main         BOOLEAN NOT NULL,
    position        INTEGER NOT NULL,
    bio_on_course   TEXT NOT NULL,
);

CREATE INDEX IF NOT EXISTS idx_course_instructor_is_main ON course_instructor(is_main);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS course_instructors;
DROP INDEX IF EXISTS idx_course_instructor_is_main;
-- +goose StatementEnd
