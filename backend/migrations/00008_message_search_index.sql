-- +goose Up
CREATE TABLE message_search_index (
    message_id UUID PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    body TEXT NOT NULL DEFAULT '',
    seq BIGINT NOT NULL,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('simple', coalesce(body, ''))) STORED,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX message_search_vector_idx ON message_search_index USING GIN (search_vector);
CREATE INDEX message_search_conv_idx ON message_search_index (conversation_id);

-- +goose Down
DROP TABLE IF EXISTS message_search_index;
