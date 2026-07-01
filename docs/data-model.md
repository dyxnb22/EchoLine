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

