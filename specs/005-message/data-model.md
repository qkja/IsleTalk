# Data Model: message-svc 会话消息服务

**Feature**: `specs/005-message` | **Date**: 2026-11-14

本文件定义 message-svc 的 2 个实体的字段、索引与状态机。领域模型与 BSON 文档解耦（宪法原则 V）。

## 实体总表

| 实体 | 集合 | 聚合根 | 关键不变量 |
|---|---|---|---|
| 会话 | `message.conversations` | ✅ | `conversation_id` 唯一；类型（首段仅 direct） |
| 消息 | `message.messages` | ✅ | `message_id` 唯一（幂等）；`(conversation_id, seq)` 唯一（时序） |

## 会话（`message.conversations`）

- 字段：`conversation_id`（唯一）、`type`（`direct` / `group`，首段仅 direct）、`participants`（参与者）、
  `last_message_preview`（最新消息摘要）、`unread_count`、`pinned`（置顶）、`muted`（免打扰）、
  `deleted`（已删除标记）、`last_active_at`（最近活跃时间）。
- 索引：`conversation_id` 唯一、`participants` 普通。
- 约束：单聊会话由好友关系派生，无需对方同意（FR-001）；删除会话后对方发新消息时重新出现（FR-009）。

## 消息（`message.messages`）

- 字段：`message_id`（幂等键，唯一）、`conversation_id`、`sender_id`、`type`（text/image/file）、
  `content`（文本正文或附件 key）、`seq`（会话内单调递增）、`sent_at`、`status`（sent/delivered/read/failed）、
  `recalled`（撤回标记）。
- 索引：`message_id` 唯一、`(conversation_id, seq)` 唯一。
- 状态机：`sent → delivered → read`；`failed` 为发送失败终态；`recalled` 为撤回标记（正文不再可见）。
- 约束：seq 由服务端原子分配，MUST 单调递增（FR-031）；撤回窗口 2 分钟由服务端判定（FR-020）。

## 跨服务关系

- message-svc **被** 客户端（发消息/拉取/撤回/已读/会话管理）调用。
- message-svc **依赖** relationship-svc：`CanMessage`（发消息前授权校验）。
- 消息 sender_id 引用 identity-svc 的 UserId；对外不暴露手机号等身份敏感信息。
