-- +goose Up
CREATE TABLE devices (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX devices_user_id_idx ON devices (user_id);

CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    type TEXT NOT NULL CHECK (type IN ('direct', 'group', 'channel')),
    title TEXT NOT NULL DEFAULT '',
    latest_seq BIGINT NOT NULL DEFAULT 0,
    last_message_id UUID,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX conversations_type_idx ON conversations (type);

CREATE TABLE conversation_members (
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'subscriber')),
    last_read_seq BIGINT NOT NULL DEFAULT 0,
    last_delivered_seq BIGINT NOT NULL DEFAULT 0,
    muted_until TIMESTAMPTZ,
    joined_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX conversation_members_user_id_idx ON conversation_members (user_id);

CREATE TABLE direct_conversation_pairs (
    user_low UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_high UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    PRIMARY KEY (user_low, user_high),
    UNIQUE (conversation_id),
    CHECK (user_low < user_high)
);

CREATE TABLE messages (
    id UUID PRIMARY KEY,
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id),
    client_msg_id TEXT NOT NULL DEFAULT '',
    seq BIGINT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('text', 'image', 'file', 'system')),
    body TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'normal' CHECK (status IN ('normal', 'edited', 'recalled', 'deleted')),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (conversation_id, seq),
    UNIQUE (sender_id, client_msg_id)
);

CREATE INDEX messages_conversation_id_seq_idx ON messages (conversation_id, seq DESC);

-- +goose Down
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS direct_conversation_pairs;
DROP TABLE IF EXISTS conversation_members;
DROP TABLE IF EXISTS conversations;
DROP TABLE IF EXISTS devices;
