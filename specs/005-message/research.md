# Research: message-svc 会话消息服务

**Feature**: `specs/005-message` | **Date**: 2026-11-14

本文件记录 message-svc 的特有决策（承接 002 的单聊消息语义；语言/存储选型已在宪法锁定）。

## 决策记录

### R-1 消息幂等去重（FR-007）

- **Decision**: `messages.message_id` 唯一索引，发送时按 message_id upsert；重发相同 message_id 返回已有消息，MUST NOT 产生重复。
- **Rationale**: FR-028 要求「重发不产生重复消息」，唯一索引是最可靠的幂等兜底（防网络重试双写）。
- **Alternatives considered**: 应用层先查后写 → 拒绝（竞态下仍可能重复）。

### R-2 会话内时序（FR-010/031）

- **Decision**: 每条消息带 `seq`（会话内单调递增序号），`(conversation_id, seq)` 唯一索引；seq 由服务端原子分配，客户端按 seq 排序保证多端顺序一致。
- **Rationale**: FR-031 要求「会话内 seq 单调递增保证时序」；唯一索引防并发发送时 seq 冲突。
- **Alternatives considered**: 仅按时间戳排序 → 拒绝（时钟偏差会导致乱序）。

### R-3 撤回时间窗（FR-006/020）

- **Decision**: 撤回窗口默认 2 分钟（可调），由服务端时钟判定；撤回校验发送者身份，非本人撤回拒绝；撤回后 `recalled` 标记，双方看到一致占位提示。
- **Rationale**: FR-020 要求「允许的时间窗内撤回」，服务端时钟判定防客户端篡改。
- **Alternatives considered**: 客户端判定窗口 → 拒绝（可被篡改）。

### R-4 未读计数（FR-023）

- **Decision**: 未读数量存会话文档（MongoDB 为持久事实），Redis 缓存热路径；MarkRead 时更新 MongoDB 并失效缓存。
- **Rationale**: Redis 非唯一事实来源（宪法 A 组）；未读计数需多端一致（FR-026）。
- **Alternatives considered**: 仅 Redis 存未读 → 拒绝（Redis 故障丢失未读状态）。

### R-5 发消息授权（FR-011）

- **Decision**: 发消息前调 relationship-svc 的 `CanMessage`，校验双方互为好友且未被拉黑；非好友或被拉黑拒绝发送。
- **Rationale**: 002 FR-013「单聊」的前提是好友关系；拉黑拦截由 relationship 承载（FR-011）。
- **Alternatives considered**: message 自行维护好友关系 → 拒绝（数据冗余、越界）。

### R-6 附件存储（FR-003/004）

- **Decision**: 图片/文件正文存对象存储（002 待办），消息仅存附件引用 key；对象存储选型确定前，用本地 mock 占位。
- **Rationale**: 消息文档不存大二进制（MongoDB 单文档上限）；附件独立存储便于 CDN 分发。
- **Alternatives considered**: 附件存 MongoDB GridFS → 拒绝（首段切片不引入，等对象存储选型）。

### R-7 会话 ID 派生（FR-001/011）

- **Decision**: 单聊会话 ID 由两个用户 ID 排序后派生（`direct:{min}:{max}`），服务端据此反解双方参与者；发消息前从会话 ID 反解接收方以调 relationship `CanMessage`。
- **Rationale**: 首段契约 `SendMessageRequest` 只携带 `sender_id` 与 `conversation_id`，不携带接收方；接收方必须由会话派生。排序保证确定性（FR-001「由好友关系派生」）。
- **Alternatives considered**: 会话 ID 为不透明 UUID → 拒绝（无 CreateConversation RPC，首条消息无法得知接收方）。
