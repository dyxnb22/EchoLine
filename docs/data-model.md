# EchoLine 数据模型

本文档记录核心表设计。PostgreSQL 是 source of truth，Redis 和搜索索引都可以从 DB 或事件流重建。

## 核心实体

### users

- `id`
- `username`
- `display_name`
- `password_hash`
- `created_at`
- `updated_at`

### devices

- `id`
- `user_id`
- `device_name`
- `platform`
- `last_seen_at`
- `created_at`

### conversations

- `id`
- `type`: `direct`、`group`、`channel`
- `title`
- `latest_seq`
- `last_message_id`
- `created_by`
- `created_at`
- `updated_at`

### conversation_members

- `conversation_id`
- `user_id`
- `role`: `owner`、`admin`、`member`、`subscriber`
- `last_read_seq`
- `last_delivered_seq`
- `muted_until`
- `archived_at` (nullable) — per-member archive（migration `00013`）
- `joined_at`

### direct_conversation_pairs

- `user_low`
- `user_high`
- `conversation_id`

用于保证两个用户之间只有一个 direct conversation。

### messages

- `id`
- `conversation_id`
- `sender_id`
- `client_msg_id`
- `seq`
- `type`: `text`、`image`、`file`、`system`
- `body`
- `status`: `normal`、`edited`、`recalled`、`deleted`
- `parent_message_id` (nullable) — thread reply parent（migration `00010`）
- `created_at`
- `updated_at`

建议唯一约束：

- `(conversation_id, seq)`
- `(sender_id, client_msg_id)`

### message_deliveries

- `message_id`
- `user_id`
- `device_id`
- `status`: `sent`、`delivered`、`read`
- `acked_at`

### attachments

- `id`
- `message_id`
- `owner_id`
- `object_key`
- `mime_type`
- `size_bytes`
- `checksum`
- `created_at`

### dead_letter_events

- `id`
- `source_topic`
- `payload`
- `error_message`
- `attempts`
- `created_at`

### device_sync_cursors

- `user_id`
- `device_id`
- `conversation_id`
- `last_seq`
- `updated_at`

### message_search_index

- `message_id`
- `conversation_id`
- `sender_id`
- `body`
- `seq`
- `search_vector` (tsvector, generated)
- `created_at`

### pinned_messages

- `conversation_id`
- `message_id`
- `pinned_by`
- `pinned_at`

### user_blocks

- `blocker_id`
- `blocked_id`
- `created_at`

### message_reports

- `id`
- `reporter_id`
- `message_id`
- `conversation_id`
- `reason`
- `created_at`

### notification_events

- `id`
- `user_id`
- `type`
- `payload`
- `read_at`
- `created_at`

### message_reactions

- `message_id`
- `user_id`
- `emoji`
- `created_at`

### push_tokens

- `id`
- `user_id`
- `device_id`
- `token`
- `platform`
- `created_at`

### payment_ledger

- `id`, `user_id`, `amount_cents`, `currency`, `status`, `reference`, `created_at`

Migration `00012`; settle via `POST /api/payments/ledger/settle` may grant channel entitlements when `reference = channel:{uuid}`.

### ad_campaigns

- `id`, `channel_id`, `title`, `status`, `created_at`
- `budget_cents`, `frequency_cap` (migration `00014`)

### ad_impressions

- `id`, `campaign_id`, `user_id`, `created_at`
- `impression_day` (migration `00014`) — daily frequency cap index

### encryption_key_bundles

- `id`, `user_id`, `device_id`, `public_key`, `created_at`
- unique `(user_id, device_id)`

Migration `00012`.

### users.is_admin

Migration `00014`: boolean admin flag (default false). Runtime admin also via `ADMIN_USER_IDS` env.

### webhook_deliveries

- `id`, `event_type`, `payload`, `status`, `attempts`, `last_error`, `created_at`, `delivered_at`

### ad_campaigns extensions (00014)

- `budget_cents`, `frequency_cap` (default 3 impressions/user/day)

### ad_impressions.impression_day (00014)

- `DATE NOT NULL DEFAULT CURRENT_DATE` — used for per-user daily frequency cap; unique index on `(campaign_id, user_id, impression_day)` (avoids non-immutable `created_at::date` in PostgreSQL indexes)

### channel_entitlements (00015)

- `user_id`, `channel_id`, `status`, `reference`, `expires_at` — paid channel access skeleton

### conversations.requires_entitlement (00016)

- `BOOLEAN NOT NULL DEFAULT FALSE` on `conversations` — when true, subscribe requires active `channel_entitlements` row

### outbox_events

Transactional outbox for reliable async publish (migration `00004`).

- `id`
- `topic` — e.g. `message.created`
- `payload` — JSONB event body
- `status`: `pending`, `processing`, `published`, `failed` (migration `00017` adds `processing` claim state)
- `attempts`
- `created_at`, `published_at`

Worker claims `pending` rows into `processing` with `SKIP LOCKED` before publish (`internal/outbox`, `cmd/worker`).

### audit_logs

- `id`
- `actor_id`
- `action`
- `resource_type`
- `resource_id`
- `metadata`
- `created_at`

## 设计原则

- conversation 内使用递增 `seq` 保证读侧顺序。
- 未读数优先用 `latest_seq - last_read_seq`，避免逐条计数。
- `client_msg_id` 用于客户端重试去重。
- 审计日志 append-only，不覆盖历史。

