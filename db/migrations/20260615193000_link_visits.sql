-- +goose Up
CREATE TABLE link_visits (
    id BIGSERIAL PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip TEXT NOT NULL,
    user_agent TEXT NOT NULL,
    referer TEXT NOT NULL DEFAULT '',
    status INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE link_visits;
