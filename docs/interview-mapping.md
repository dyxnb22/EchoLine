# EchoLine 面试题映射

| 面试主题 | 项目模块 | 真实工程问题 | 面试常见问法 | 做完后的讲述点 |
|---|---|---|---|---|
| WebSocket 长连接 | `realtime` | 连接管理、心跳、断线重连 | 如何设计百万长连接？ | 网关维护连接，业务写 DB，推送解耦 |
| 消息可靠性 | `message` | 写入成功但推送失败 | 如何保证消息不丢？ | 先持久化，再投递；客户端可拉取补偿 |
| ACK / 重试 / 去重 | `delivery` | 重复发送、网络抖动 | 客户端重试导致重复怎么办？ | `client_msg_id` + unique key + ACK 状态 |
| 顺序性 | `message` | 同会话消息乱序 | 如何保证聊天消息有序？ | conversation 内递增 seq，读侧按 seq 排序 |
| 会话列表 | `conversation` | 最近消息、排序、置顶 | 微信会话列表怎么做？ | conversation member 保存 last_read / last_delivered |
| 未读数 | `conversation` | 多端读状态同步 | 未读数如何准确计算？ | `latest_seq - last_read_seq`，避免逐条计数 |
| 离线消息 | `sync` | 用户离线后补消息 | 用户离线一天如何同步？ | 按 device cursor / conversation seq 增量拉取 |
| 历史同步 | `message` | 分页、游标、冷热数据 | 历史消息怎么分页？ | cursor pagination，按 conversation_id + seq 索引 |
| 群聊扩散 | `fanout` | 大群写扩散成本高 | 群消息用读扩散还是写扩散？ | 小群写扩散，大群读扩散或混合策略 |
| 频道模型 | `channel` | 广播给大量订阅者 | 频道和群聊区别？ | 频道是发布订阅，弱互动，适合读扩散 |
| 多端同步 | `device` | 手机、Web 同时在线 | 多端消息状态如何一致？ | device session + per-device ack + account-level read |
| 在线状态 | `presence` | 心跳、过期、雪崩 | 在线状态怎么存？ | Redis TTL + heartbeat + 最终一致 |
| Redis 与 DB 配合 | `cache` | 缓存穿透、脏读 | Redis 和 DB 如何配合？ | DB 为准，Redis 缓存热点会话和 presence |
| MQ 解耦 | `eventbus` | 同步链路过重 | 为什么需要 MQ？ | 消息持久化后发事件，通知、搜索、审计异步消费 |
| 限流 | `rate_limit` | 刷屏、接口滥用 | 如何防刷消息？ | 用户、会话、IP 维度令牌桶 |
| 风控 | `risk` | 垃圾消息、异常行为 | IM 如何做基础风控？ | 频率、重复内容、拉黑、审计事件 |
| 热点群 / 高并发 | `hot_conversation` | 大群消息风暴 | 大群 10 万人怎么推？ | 分层推送、异步 fanout、在线优先、离线拉取 |
| 搜索 | `search` | 消息全文检索延迟 | 聊天记录搜索怎么做？ | OpenSearch 异步建索引，DB 为源数据 |
| 附件链路 | `media` | 大文件、鉴权、下载 | 图片视频怎么传？ | 预签名 URL，元数据入库，异步安全检查 |
| 审计日志 | `audit` | 操作追溯 | 撤回和删除如何审计？ | append-only audit log，业务表只保留当前状态 |
| 可观测性 | `observability` | 排查延迟和丢消息 | 消息慢在哪里怎么查？ | trace_id 串联 API、DB、MQ、WS delivery |
| 扩展性 / 分库分表 | `docs/scaling` | 单库瓶颈 | 消息表太大怎么办？ | 按 conversation_id 分片，冷热分层，归档策略 |

